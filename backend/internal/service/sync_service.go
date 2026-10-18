package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"roadside-league/backend/internal/ddragon"
	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/storage"
)

// SyncState 同步结果状态。
type SyncState struct {
	Running    bool   `json:"running"`
	Status     string `json:"status"` // idle / success / failed
	Version    string `json:"version"`
	Stage      string `json:"stage"`
	Total      int    `json:"total"`
	Done       int    `json:"done"`
	StageTotal int    `json:"stageTotal"`
	StageDone  int    `json:"stageDone"`
	Failed     int    `json:"failed"`
	Downloaded int    `json:"downloaded"`
	Message    string `json:"message"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
}

type SyncService struct {
	local    *ddragon.LocalReader
	heroes   *repository.HeroRepository
	assets   *repository.AssetRepository
	runes    *repository.RuneRepository
	settings *repository.SettingRepository
	store    *storage.LocalStore

	mu    sync.Mutex
	state SyncState
}

func NewSyncService(
	local *ddragon.LocalReader,
	heroes *repository.HeroRepository,
	assets *repository.AssetRepository,
	runes *repository.RuneRepository,
	settings *repository.SettingRepository,
	store *storage.LocalStore,
) *SyncService {
	return &SyncService{
		local: local, heroes: heroes, assets: assets, runes: runes,
		settings: settings, store: store,
	}
}

// State 返回当前同步状态。
func (s *SyncService) State() SyncState {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.state
	if st.Version == "" {
		st.Version, _ = s.settings.Get(model.SettingDDragonVersion)
	}
	if st.Status == "" {
		st.Status = "idle"
		if v, _ := s.settings.Get(model.SettingDDragonLastSync); v != "" {
			st.FinishedAt = v
		}
	}
	return st
}

// Run 从本地 assets 离线同步数据到数据库与媒体目录。
// 在启动时同步调用；若版本未变则跳过。
func (s *SyncService) Run() error {
	s.mu.Lock()
	s.state = SyncState{Running: true, Status: "running", Stage: "准备中", StartedAt: nowStr()}
	s.mu.Unlock()

	version := s.local.Version()
	s.mu.Lock()
	s.state.Version = version
	s.mu.Unlock()

	// 版本未变则跳过（meta 校验通过即视为一致）
	if current, _ := s.settings.Get(model.SettingDDragonVersion); current == version && current != "" {
		s.finish("skipped", "当前已是最新版本 "+version)
		return nil
	}

	// 1. 读取 JSON 数据
	s.setStage("读取本地数据", 0)
	champs, err := s.local.Champions()
	if err != nil {
		s.finish("failed", err.Error())
		return err
	}
	items, err := s.local.Items()
	if err != nil {
		s.finish("failed", err.Error())
		return err
	}
	summoners, err := s.local.Summoners()
	if err != nil {
		s.finish("failed", err.Error())
		return err
	}
	runeTrees, err := s.local.Runes()
	if err != nil {
		s.finish("failed", err.Error())
		return err
	}

	var failedTotal int64

	// 2. 英雄
	heroJobs := make([]iconJob, 0, len(champs.Data))
	for _, ch := range champs.Data {
		heroJobs = append(heroJobs, iconJob{
			dir: "heroes", file: ch.Image.Full,
			url: "/ddragon/champion/" + ch.Image.Full,
			apply: func(url string) error {
				return s.heroes.UpsertDDragon(&model.Hero{
					Name: ch.Name, Title: ch.Title, Lane: "",
					DDragonID: ch.ID, DDragonKey: ch.Key,
					Aliases: model.JoinList(ddragon.HeroAliases[ch.ID]),
					Roles:   model.JoinList(ddragon.RolesFromTags(ch.Tags)),
					Avatar:  url,
				})
			},
		})
	}
	failedTotal += int64(s.runJobs("同步英雄图标", heroJobs))

	// 3. 装备
	itemJobs := make([]iconJob, 0, len(items.Data))
	for id, item := range items.Data {
		if !summonerRiftItem(id, item) {
			continue
		}
		file := id + ".png"
		sub := itemSubCategory(item.Tags)
		itemJobs = append(itemJobs, iconJob{
			dir: "items", file: file,
			url: "/ddragon/item/" + file,
			apply: func(url string) error {
				return s.assets.UpsertDDragon(&model.GameAsset{
					Category: model.AssetCategoryItem, SubCategory: sub,
					Name: item.Name, DDragonID: id, Version: version, Icon: url,
				})
			},
		})
	}
	failedTotal += int64(s.runJobs("同步装备图标", itemJobs))

	// 4. 召唤师技能
	spellJobs := make([]iconJob, 0)
	for _, sp := range summoners.Data {
		if !containsStr(sp.Modes, "CLASSIC") || sp.Image.Full == "" {
			continue
		}
		spellJobs = append(spellJobs, iconJob{
			dir: "spells", file: sp.Image.Full,
			url: "/ddragon/spell/" + sp.Image.Full,
			apply: func(url string) error {
				return s.assets.UpsertDDragon(&model.GameAsset{
					Category: model.AssetCategorySkill, SubCategory: "召唤师技能",
					Name: sp.Name, DDragonID: sp.ID, Version: version, Icon: url,
				})
			},
		})
	}
	failedTotal += int64(s.runJobs("同步召唤师技能图标", spellJobs))

	// 5. 符文
	runeJobs := make([]iconJob, 0)
	for _, tree := range runeTrees {
		treeID := int64(tree.ID)
		treeName, treeKey, treeIcon := tree.Name, tree.Key, tree.Icon
		treeSort := ddragon.TreeSort(treeID)
		treeFile := strconv.Itoa(tree.ID) + ".png"
		runeJobs = append(runeJobs, iconJob{
			dir: "runes", file: treeFile,
			url: "/ddragon/" + treeIcon,
			apply: func(url string) error {
				if err := s.assets.UpsertDDragon(&model.GameAsset{
					Category: model.AssetCategoryRune, SubCategory: treeName,
					Name: treeName, DDragonID: strconv.Itoa(tree.ID), Version: version, Icon: url,
				}); err != nil {
					return err
				}
				return s.runes.UpsertTree(&model.RuneTree{
					ID: treeID, Key: treeKey, Name: treeName, Icon: url, Sort: treeSort,
				})
			},
		})
		for slotIdx, slot := range tree.Slots {
			row := slotIdx
			isKeystone := slotIdx == 0
			for runeIdx, rune_ := range slot.Runes {
				runeID := int64(rune_.ID)
				runeName, runeIcon := rune_.Name, rune_.Icon
				runeSort := runeIdx
				runeFile := strconv.Itoa(rune_.ID) + ".png"
				runeJobs = append(runeJobs, iconJob{
					dir: "runes", file: runeFile,
					url: "/ddragon/" + runeIcon,
					apply: func(url string) error {
						if err := s.assets.UpsertDDragon(&model.GameAsset{
							Category: model.AssetCategoryRune, SubCategory: treeName,
							Name: runeName, DDragonID: strconv.Itoa(rune_.ID), Version: version, Icon: url,
						}); err != nil {
							return err
						}
						return s.runes.UpsertRune(&model.Rune{
							ID: runeID, TreeID: treeID, Name: runeName, Icon: url,
							Row: row, IsKeystone: isKeystone, Sort: runeSort,
						})
					},
				})
			}
		}
	}
	failedTotal += int64(s.runJobs("同步符文图标", runeJobs))

	// 6. 属性碎片（仅落盘）
	shardJobs := make([]iconJob, 0, 6)
	for _, f := range ddragon.ShardFiles() {
		file := f + ".png"
		shardJobs = append(shardJobs, iconJob{
			dir: "shards", file: file,
			url:   "/ddragon/statmods/" + file,
			apply: func(string) error { return nil },
		})
	}
	failedTotal += int64(s.runJobs("同步属性碎片", shardJobs))

	// 7. 记录版本
	if failedTotal > 0 {
		s.finish("success", fmt.Sprintf("版本 %s 同步完成，%d 个图标失败", version, failedTotal))
		return nil
	}
	if err := s.settings.Set(model.SettingDDragonVersion, version); err != nil {
		s.finish("failed", "写入同步版本失败: "+err.Error())
		return err
	}
	if err := s.settings.Set(model.SettingDDragonLastSync, nowStr()); err != nil {
		s.finish("failed", "写入同步时间失败: "+err.Error())
		return err
	}

	st := s.State()
	s.finish("success", fmt.Sprintf("已同步到版本 %s（%d 项，新复制 %d 个图标）", version, st.Done, st.Downloaded))
	return nil
}

func (s *SyncService) finish(status, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Running = false
	s.state.Status = status
	s.state.Message = message
	s.state.Stage = ""
	s.state.FinishedAt = nowStr()
}

func (s *SyncService) setStage(stage string, total int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Stage = stage
	s.state.Total += total
	s.state.StageTotal = total
	s.state.StageDone = 0
}

func (s *SyncService) tick(copied bool, failed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Done++
	s.state.StageDone++
	if copied {
		s.state.Downloaded++
	}
	if failed {
		s.state.Failed++
	}
}

// ---------- 并发执行 ----------

type iconJob struct {
	dir   string
	file  string
	url   string // 直接指向 assets 的对外 URL（/ddragon/...），无需复制到 media
	apply func(url string) error
}

func (s *SyncService) runJobs(stage string, jobs []iconJob) int {
	s.setStage(stage, len(jobs))
	var wg sync.WaitGroup
	var failCount int64
	sem := make(chan struct{}, 8)

	for _, job := range jobs {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			failed := false
			if err := job.apply(job.url); err != nil {
				log.Printf("sync upsert failed [%s/%s]: %v", job.dir, job.file, err)
				failed = true
				failCount++
			}
			s.tick(false, failed)
		}()
	}
	wg.Wait()
	return int(failCount)
}

// ---------- 装备筛选与分类 ----------

func summonerRiftItem(id string, item ddragon.Item) bool {
	if item.Name == "" || item.Image.Full == "" {
		return false
	}
	if !item.Maps["11"] {
		return false
	}
	if item.InStore != nil && !*item.InStore {
		return false
	}
	if item.Gold.Purchasable != nil && !*item.Gold.Purchasable {
		return false
	}
	if id == "0" {
		return false
	}
	return true
}

func itemSubCategory(tags []string) string {
	has := func(want string) bool { return containsStr(tags, want) }
	switch {
	case has("Consumable"):
		return "消耗"
	case has("Boots"):
		return "移动"
	case has("Damage"), has("CriticalStrike"), has("AttackSpeed"),
		has("LifeSteal"), has("OnHit"), has("ArmorPenetration"):
		return "攻击"
	case has("SpellDamage"), has("ManaRegeneration"), has("Mana"),
		has("MagicPenetration"), has("CooldownReduction"), has("AbilityHaste"):
		return "法术"
	case has("Health"), has("Armor"), has("SpellBlock"), has("HealthRegeneration"):
		return "防御"
	default:
		return model.ItemSubOther
	}
}

func containsStr(list []string, target string) bool {
	for _, v := range list {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}

func nowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
