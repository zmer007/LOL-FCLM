package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"roadside-league/backend/internal/ddragon"
	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
	"roadside-league/backend/internal/storage"
)

var ErrSyncRunning = errors.New("sync already running")

// 同步状态（也直接作为接口返回 JSON）
type SyncState struct {
	Running    bool   `json:"running"`
	Status     string `json:"status"` // idle / running / success / skipped / failed
	Version    string `json:"version"`
	Stage      string `json:"stage"`
	Total      int    `json:"total"`      // 全部作业数
	Done       int    `json:"done"`       // 全部完成数（含失败）
	StageTotal int    `json:"stageTotal"` // 当前阶段总数
	StageDone  int    `json:"stageDone"`  // 当前阶段完成数
	Failed     int    `json:"failed"`
	Downloaded int    `json:"downloaded"`
	Message    string `json:"message"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
}

type SyncService struct {
	cli      *ddragon.Client
	heroes   *repository.HeroRepository
	assets   *repository.AssetRepository
	settings *repository.SettingRepository
	store    *storage.LocalStore

	mu    sync.Mutex
	state SyncState
}

func NewSyncService(
	cli *ddragon.Client,
	heroes *repository.HeroRepository,
	assets *repository.AssetRepository,
	settings *repository.SettingRepository,
	store *storage.LocalStore,
) *SyncService {
	return &SyncService{cli: cli, heroes: heroes, assets: assets, settings: settings, store: store}
}

// State 返回当前同步状态（附带上次成功版本与时间）。
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

// Start 异步触发一次同步；已在运行时返回 ErrSyncRunning。
func (s *SyncService) Start(force bool) error {
	s.mu.Lock()
	if s.state.Running {
		s.mu.Unlock()
		return ErrSyncRunning
	}
	s.state = SyncState{Running: true, Status: "running", Stage: "准备中", StartedAt: nowStr(), Version: s.state.Version}
	s.mu.Unlock()

	go func() {
		if err := s.run(force); err != nil {
			s.finish("failed", err.Error())
		}
	}()
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

func (s *SyncService) tick(downloaded bool, failed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Done++
	s.state.StageDone++
	if downloaded {
		s.state.Downloaded++
	}
	if failed {
		s.state.Failed++
	}
}

func (s *SyncService) run(force bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// 1. 最新版本号
	version, err := s.cli.LatestVersion(ctx)
	if err != nil {
		return fmt.Errorf("获取最新版本失败: %w", err)
	}
	s.mu.Lock()
	s.state.Version = version
	s.mu.Unlock()

	if !force {
		if current, _ := s.settings.Get(model.SettingDDragonVersion); current == version && current != "" {
			s.finish("skipped", "当前已是最新版本 "+version)
			return nil
		}
	}

	// 2. 拉取四类 JSON
	s.setStage("获取官方数据", 0)
	champs, err := s.cli.Champions(ctx, version)
	if err != nil {
		return fmt.Errorf("获取英雄数据失败: %w", err)
	}
	items, err := s.cli.Items(ctx, version)
	if err != nil {
		return fmt.Errorf("获取装备数据失败: %w", err)
	}
	summoners, err := s.cli.Summoners(ctx, version)
	if err != nil {
		return fmt.Errorf("获取召唤师技能数据失败: %w", err)
	}
	runeTrees, err := s.cli.Runes(ctx, version)
	if err != nil {
		return fmt.Errorf("获取符文数据失败: %w", err)
	}

	var failedTotal int64

	// 3. 英雄
	heroJobs := make([]iconJob, 0, len(champs.Data))
	for _, ch := range champs.Data {
		heroJobs = append(heroJobs, iconJob{
			dir: "heroes", file: ch.Image.Full,
			remote: "img/champion/" + ch.Image.Full,
			apply: func(url string) error {
				return s.heroes.UpsertDDragon(&model.Hero{
					Name: ch.Name, Title: ch.Title, Lane: "",
					DDragonID: ch.ID, DDragonKey: ch.Key, Avatar: url,
				})
			},
		})
	}
	failedTotal += int64(s.runJobs(ctx, "同步英雄图标", heroJobs))

	// 4. 装备（仅召唤师峡谷可购买）
	itemJobs := make([]iconJob, 0, len(items.Data))
	for id, item := range items.Data {
		if !summonerRiftItem(id, item) {
			continue
		}
		file := id + ".png"
		sub := itemSubCategory(item.Tags)
		itemJobs = append(itemJobs, iconJob{
			dir: "items", file: file,
			remote: "img/item/" + file,
			apply: func(url string) error {
				return s.assets.UpsertDDragon(&model.GameAsset{
					Category: model.AssetCategoryItem, SubCategory: sub,
					Name: item.Name, DDragonID: id, Version: version, Icon: url,
				})
			},
		})
	}
	failedTotal += int64(s.runJobs(ctx, "同步装备图标", itemJobs))

	// 5. 召唤师技能（仅经典峡谷模式）
	spellJobs := make([]iconJob, 0)
	for _, sp := range summoners.Data {
		if !containsStr(sp.Modes, "CLASSIC") || sp.Image.Full == "" {
			continue
		}
		spellJobs = append(spellJobs, iconJob{
			dir: "spells", file: sp.Image.Full,
			remote: "img/spell/" + sp.Image.Full,
			apply: func(url string) error {
				return s.assets.UpsertDDragon(&model.GameAsset{
					Category: model.AssetCategorySkill, SubCategory: "召唤师技能",
					Name: sp.Name, DDragonID: sp.ID, Version: version, Icon: url,
				})
			},
		})
	}
	failedTotal += int64(s.runJobs(ctx, "同步召唤师技能图标", spellJobs))

	// 6. 符文：每个符文系本身 + 各排符文（用数字 id 命名避免重名）
	runeJobs := make([]iconJob, 0)
	for _, tree := range runeTrees {
		treeFile := strconv.Itoa(tree.ID) + ".png"
		runeJobs = append(runeJobs, iconJob{
			dir: "runes", file: treeFile, global: true,
			remote: "img/" + tree.Icon,
			apply: func(url string) error {
				return s.assets.UpsertDDragon(&model.GameAsset{
					Category: model.AssetCategoryRune, SubCategory: tree.Name,
					Name: tree.Name, DDragonID: strconv.Itoa(tree.ID), Version: version, Icon: url,
				})
			},
		})
		for _, slot := range tree.Slots {
			for _, rune_ := range slot.Runes {
				runeFile := strconv.Itoa(rune_.ID) + ".png"
				runeJobs = append(runeJobs, iconJob{
					dir: "runes", file: runeFile, global: true,
					remote: "img/" + rune_.Icon,
					apply: func(url string) error {
						return s.assets.UpsertDDragon(&model.GameAsset{
							Category: model.AssetCategoryRune, SubCategory: tree.Name,
							Name: rune_.Name, DDragonID: strconv.Itoa(rune_.ID), Version: version, Icon: url,
						})
					},
				})
			}
		}
	}
	failedTotal += int64(s.runJobs(ctx, "同步符文图标", runeJobs))

	// 7. 全部成功才记录版本（有失败时下次点击仍会重试缺失图标）
	if failedTotal > 0 {
		s.finish("success", fmt.Sprintf("版本 %s 同步完成，%d 个图标失败，可再次点击更新重试", version, failedTotal))
		return nil
	}
	if err := s.settings.Set(model.SettingDDragonVersion, version); err != nil {
		return fmt.Errorf("写入同步版本失败: %w", err)
	}
	if err := s.settings.Set(model.SettingDDragonLastSync, nowStr()); err != nil {
		return fmt.Errorf("写入同步时间失败: %w", err)
	}

	st := s.State()
	s.finish("success", fmt.Sprintf("已同步到版本 %s（%d 项，新下载 %d 个图标）", version, st.Done, st.Downloaded))
	return nil
}

// ---------- 并发执行 ----------

type iconJob struct {
	dir    string
	file   string
	remote string                      // CDN 上的相对路径
	global bool                        // true 表示与版本无关（cdn/img/，符文图标）
	apply  func(localURL string) error // 落盘后 upsert 数据库
}

func (s *SyncService) runJobs(ctx context.Context, stage string, jobs []iconJob) int {
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

			localURL := s.store.IconURL(job.dir, job.file)
			downloaded := false
			failed := false

			if !s.store.IconExists(job.dir, job.file) {
				var body io.ReadCloser
				var err error
				if job.global {
					body, err = s.cli.OpenIconGlobal(ctx, job.remote)
				} else {
					body, err = s.cli.OpenIcon(ctx, s.State().Version, job.remote)
				}
				if err == nil {
					url, wrote, saveErr := s.store.SaveIconFile(job.dir, job.file, body)
					body.Close()
					if saveErr == nil {
						localURL, downloaded = url, wrote
					} else {
						err = saveErr
					}
				}
				if err != nil {
					log.Printf("sync icon failed [%s/%s] %s: %v", job.dir, job.file, job.remote, err)
					failed = true
					atomic.AddInt64(&failCount, 1)
				}
			}

			if !failed {
				if err := job.apply(localURL); err != nil {
					log.Printf("sync upsert failed [%s/%s]: %v", job.dir, job.file, err)
					atomic.AddInt64(&failCount, 1)
					failed = true
				}
			}
			s.tick(downloaded, failed)
		}()
	}
	wg.Wait()
	return int(atomic.LoadInt64(&failCount))
}

// ---------- 装备筛选与分类 ----------

func summonerRiftItem(id string, item ddragon.Item) bool {
	if item.Name == "" || item.Image.Full == "" {
		return false
	}
	if !item.Maps["11"] { // 11 = 召唤师峡谷
		return false
	}
	if item.InStore != nil && !*item.InStore {
		return false
	}
	if item.Gold.Purchasable != nil && !*item.Gold.Purchasable {
		return false
	}
	// 排除纯数字异常与基础占位（如 0 号物品）
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
