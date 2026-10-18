// ddragon-fetch 从 Riot Data Dragon 下载指定版本数据到 assets 目录，
// 生成 meta.json 元数据供离线部署使用。
//
// 用法: go run ./cmd/ddragon-fetch <version> [assets-root]
//
//	version: 游戏版本号，如 16.20.1；latest 表示自动获取最新版本
//	assets-root: 可选，默认为项目根目录下的 assets
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"roadside-league/backend/internal/ddragon"
)

const (
	ddragonBase = "https://ddragon.leagueoflegends.com/cdn"
	versionsURL = "https://ddragon.leagueoflegends.com/api/versions.json"
	cdragonBase = "https://raw.communitydragon.org/latest/plugins/rcp-be-lol-game-data/global/default"
	defaultLang = "zh_CN"
	httpTimeout = 60 * time.Second
)

var httpClient = &http.Client{Timeout: httpTimeout}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "用法: ddragon-fetch <version|latest> [assets-root]")
		os.Exit(1)
	}
	version := os.Args[1]
	assetsRoot := "assets"
	if len(os.Args) >= 3 {
		assetsRoot = os.Args[2]
	}

	ctx := context.Background()

	if version == "latest" {
		v, err := latestVersion(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取最新版本失败: %v\n", err)
			os.Exit(1)
		}
		version = v
		fmt.Printf("最新版本: %s\n", version)
	}

	dirName := ddragon.VersionDirName(version)
	dir := filepath.Join(assetsRoot, dirName)
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "创建目录失败: %v\n", err)
		os.Exit(1)
	}

	meta := &ddragon.Meta{
		Version:     version,
		DirName:     dirName,
		GeneratedAt: time.Now().Format(time.RFC3339),
		Lang:        defaultLang,
		Checksums:   map[string]string{},
	}

	// 1. 下载 JSON 数据
	fmt.Println("下载 champion.json...")
	champs, err := downloadJSON(ctx, version, "champion.json", new(ddragon.ChampionList), dir, meta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "下载 champion.json 失败: %v\n", err)
		os.Exit(1)
	}
	meta.Counts.Champions = len(champs.Data)
	for id, ch := range champs.Data {
		meta.Heroes = append(meta.Heroes, ddragon.MetaRef{ID: id, Key: ch.Key, Name: ch.Name})
	}

	fmt.Println("下载 item.json...")
	items, err := downloadJSON(ctx, version, "item.json", new(ddragon.ItemList), dir, meta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "下载 item.json 失败: %v\n", err)
		os.Exit(1)
	}
	meta.Counts.Items = len(items.Data)
	for id, it := range items.Data {
		meta.Items = append(meta.Items, ddragon.MetaRef{ID: id, Name: it.Name})
	}

	fmt.Println("下载 summoner.json...")
	summoners, err := downloadJSON(ctx, version, "summoner.json", new(ddragon.SummonerList), dir, meta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "下载 summoner.json 失败: %v\n", err)
		os.Exit(1)
	}
	meta.Counts.Summoners = len(summoners.Data)
	for id, sp := range summoners.Data {
		meta.Spells = append(meta.Spells, ddragon.MetaRef{ID: id, Name: sp.Name})
	}

	fmt.Println("下载 runesReforged.json...")
	runeTrees, err := downloadJSON(ctx, version, "runesReforged.json", new([]ddragon.RuneTree), dir, meta)
	if err != nil {
		fmt.Fprintf(os.Stderr, "下载 runesReforged.json 失败: %v\n", err)
		os.Exit(1)
	}
	meta.Counts.RuneTrees = len(*runeTrees)
	for _, tree := range *runeTrees {
		meta.Runes = append(meta.Runes, ddragon.MetaRef{ID: fmt.Sprint(tree.ID), Name: tree.Name})
		meta.Counts.Runes++
		for _, slot := range tree.Slots {
			for _, r := range slot.Runes {
				meta.Runes = append(meta.Runes, ddragon.MetaRef{ID: fmt.Sprint(r.ID), Name: r.Name})
				meta.Counts.Runes++
			}
		}
	}

	// 排序 meta 引用以便稳定比对
	sort.Slice(meta.Heroes, func(i, j int) bool { return meta.Heroes[i].ID < meta.Heroes[j].ID })
	sort.Slice(meta.Items, func(i, j int) bool { return meta.Items[i].ID < meta.Items[j].ID })
	sort.Slice(meta.Spells, func(i, j int) bool { return meta.Spells[i].ID < meta.Spells[j].ID })
	sort.Slice(meta.Runes, func(i, j int) bool { return meta.Runes[i].ID < meta.Runes[j].ID })

	// 2. 下载英雄图标
	fmt.Println("下载英雄图标...")
	imgDir := filepath.Join(dir, "img")
	os.MkdirAll(filepath.Join(imgDir, "champion"), 0o755)
	for _, ch := range champs.Data {
		if err := downloadIcon(ctx, version, "img/champion/"+ch.Image.Full, filepath.Join(imgDir, "champion", ch.Image.Full)); err != nil {
			fmt.Fprintf(os.Stderr, "下载英雄图标 %s 失败: %v\n", ch.Image.Full, err)
		}
	}

	// 3. 下载装备图标（仅召唤师峡谷可购买）
	fmt.Println("下载装备图标...")
	os.MkdirAll(filepath.Join(imgDir, "item"), 0o755)
	for id, item := range items.Data {
		if !summonerRiftItem(id, item) {
			continue
		}
		file := id + ".png"
		if err := downloadIcon(ctx, version, "img/item/"+file, filepath.Join(imgDir, "item", file)); err != nil {
			fmt.Fprintf(os.Stderr, "下载装备图标 %s 失败: %v\n", file, err)
		}
	}

	// 4. 下载召唤师技能图标
	fmt.Println("下载召唤师技能图标...")
	os.MkdirAll(filepath.Join(imgDir, "spell"), 0o755)
	for _, sp := range summoners.Data {
		if !containsStr(sp.Modes, "CLASSIC") || sp.Image.Full == "" {
			continue
		}
		if err := downloadIcon(ctx, version, "img/spell/"+sp.Image.Full, filepath.Join(imgDir, "spell", sp.Image.Full)); err != nil {
			fmt.Fprintf(os.Stderr, "下载召唤师技能图标 %s 失败: %v\n", sp.Image.Full, err)
		}
	}

	// 5. 下载符文图标（全局，与版本无关，CDN 路径为 img/perk-images/...）
	fmt.Println("下载符文图标...")
	for _, tree := range *runeTrees {
		downloadGlobalIcon(ctx, "img/"+tree.Icon, filepath.Join(imgDir, filepath.FromSlash(tree.Icon)))
		for _, slot := range tree.Slots {
			for _, r := range slot.Runes {
				downloadGlobalIcon(ctx, "img/"+r.Icon, filepath.Join(imgDir, filepath.FromSlash(r.Icon)))
			}
		}
	}

	// 6. 下载属性碎片图标（Community Dragon）
	fmt.Println("下载属性碎片图标...")
	os.MkdirAll(filepath.Join(imgDir, "statmods"), 0o755)
	for _, f := range ddragon.ShardFiles() {
		url := fmt.Sprintf("%s/v1/perk-images/statmods/%s.png", cdragonBase, f)
		dst := filepath.Join(imgDir, "statmods", f+".png")
		if err := downloadFile(ctx, url, dst); err != nil {
			fmt.Fprintf(os.Stderr, "下载属性碎片 %s 失败: %v\n", f, err)
		}
	}

	// 7. 写入 meta.json
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "序列化 meta.json 失败: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(dir, "meta.json"), metaData, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "写入 meta.json 失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n完成: %s 版本数据已下载到 %s\n", version, dir)
	fmt.Printf("  英雄: %d, 装备: %d, 召唤师技能: %d, 符文系: %d, 符文: %d\n",
		meta.Counts.Champions, meta.Counts.Items, meta.Counts.Summoners, meta.Counts.RuneTrees, meta.Counts.Runes)
}

// ---------- 辅助函数 ----------

func latestVersion(ctx context.Context) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, versionsURL, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var versions []string
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("versions.json 为空")
	}
	return versions[0], nil
}

func downloadJSON[T any](ctx context.Context, version, filename string, out T, dir string, meta *ddragon.Meta) (T, error) {
	url := fmt.Sprintf("%s/%s/data/%s/%s", ddragonBase, version, defaultLang, filename)
	rel := filepath.ToSlash(filepath.Join("data", filename))
	dst := filepath.Join(dir, "data", filename)
	if err := downloadFile(ctx, url, dst); err != nil {
		return out, err
	}
	data, err := os.ReadFile(dst)
	if err != nil {
		return out, err
	}
	if err := json.Unmarshal(data, out); err != nil {
		return out, err
	}
	sum := sha256.Sum256(data)
	meta.Checksums[rel] = hex.EncodeToString(sum[:])
	return out, nil
}

func downloadIcon(ctx context.Context, version, remote, dst string) error {
	url := fmt.Sprintf("%s/%s/%s", ddragonBase, version, remote)
	return downloadFile(ctx, url, dst)
}

func downloadGlobalIcon(ctx context.Context, remote, dst string) error {
	url := fmt.Sprintf("%s/%s", ddragonBase, remote)
	return downloadFile(ctx, url, dst)
}

func downloadFile(ctx context.Context, url, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		return nil // 已存在，跳过
	}
	os.MkdirAll(filepath.Dir(dst), 0o755)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

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

func containsStr(list []string, target string) bool {
	for _, v := range list {
		if strings.EqualFold(v, target) {
			return true
		}
	}
	return false
}
