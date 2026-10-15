// Package ddragon 封装 Riot Data Dragon 静态数据访问（无需 API key）。
// 文档：https://developer.riotgames.com/docs/lol#data-dragon
package ddragon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	BaseURL        = "https://ddragon.leagueoflegends.com/cdn"
	VersionsURL    = "https://ddragon.leagueoflegends.com/api/versions.json"
	DefaultLang    = "zh_CN"
	requestTimeout = 30 * time.Second
)

type Client struct {
	http *http.Client
	lang string
}

func New() *Client {
	return &Client{
		http: &http.Client{Timeout: requestTimeout},
		lang: DefaultLang,
	}
}

// LatestVersion 取当前最新游戏版本（versions.json 的第一个元素）。
func (c *Client) LatestVersion(ctx context.Context) (string, error) {
	var versions []string
	if err := c.getJSON(ctx, VersionsURL, &versions); err != nil {
		return "", err
	}
	if len(versions) == 0 {
		return "", fmt.Errorf("ddragon versions.json 为空")
	}
	return versions[0], nil
}

// ---------- 原始数据结构（仅声明需要的字段） ----------

type ChampionList struct {
	Data map[string]Champion `json:"data"`
}

type Champion struct {
	ID    string   `json:"id"`    // 英文 id，如 Ahri（也是图标文件名）
	Key   string   `json:"key"`   // 数字 key，如 103
	Name  string   `json:"name"`  // 中文名，如 九尾妖狐
	Title string   `json:"title"` // 称号，如 阿狸
	Tags  []string `json:"tags"`
	Image Image    `json:"image"`
}

type ItemList struct {
	Data map[string]Item `json:"data"`
}

type Item struct {
	Name    string          `json:"name"`
	Tags    []string        `json:"tags"`
	InStore *bool           `json:"inStore"` // 缺省视为可购买
	Maps    map[string]bool `json:"maps"`
	Gold    ItemGold        `json:"gold"`
	Image   Image           `json:"image"`
}

type ItemGold struct {
	Purchasable *bool `json:"purchasable"` // 缺省视为可购买
}

type SummonerList struct {
	Data map[string]Summoner `json:"data"`
}

type Summoner struct {
	ID    string   `json:"id"` // SummonerFlash
	Name  string   `json:"name"`
	Modes []string `json:"modes"`
	Image Image    `json:"image"`
}

// RuneTree 一个符文系（含基石与三排小符文）。
type RuneTree struct {
	ID    int        `json:"id"` // 如 8100 主宰
	Key   string     `json:"key"`
	Name  string     `json:"name"`
	Icon  string     `json:"icon"` // perk-images/Styles/...
	Slots []RuneSlot `json:"slots"`
}

type RuneSlot struct {
	Runes []Rune `json:"runes"`
}

type Rune struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type Image struct {
	Full string `json:"full"`
}

// ---------- 数据拉取 ----------

func (c *Client) Champions(ctx context.Context, version string) (*ChampionList, error) {
	var list ChampionList
	url := fmt.Sprintf("%s/%s/data/%s/champion.json", BaseURL, version, c.lang)
	if err := c.getJSON(ctx, url, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) Items(ctx context.Context, version string) (*ItemList, error) {
	var list ItemList
	url := fmt.Sprintf("%s/%s/data/%s/item.json", BaseURL, version, c.lang)
	if err := c.getJSON(ctx, url, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) Summoners(ctx context.Context, version string) (*SummonerList, error) {
	var list SummonerList
	url := fmt.Sprintf("%s/%s/data/%s/summoner.json", BaseURL, version, c.lang)
	if err := c.getJSON(ctx, url, &list); err != nil {
		return nil, err
	}
	return &list, nil
}

func (c *Client) Runes(ctx context.Context, version string) ([]RuneTree, error) {
	var list []RuneTree
	url := fmt.Sprintf("%s/%s/data/%s/runesReforged.json", BaseURL, version, c.lang)
	if err := c.getJSON(ctx, url, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// OpenIcon 打开版本相关图标流，remotePath 形如 img/champion/Ahri.png。
func (c *Client) OpenIcon(ctx context.Context, version, remotePath string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/%s", BaseURL, version, remotePath)
	return c.open(ctx, url)
}

// OpenIconGlobal 打开与版本无关的全局图标流（符文图标位于 cdn/img/ 下）。
func (c *Client) OpenIconGlobal(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s", BaseURL, remotePath)
	return c.open(ctx, url)
}

func (c *Client) open(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("下载图标失败 %s: HTTP %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}

func (c *Client) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("请求 %s 失败: HTTP %d", url, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
