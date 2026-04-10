package xinzhi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

type apiResponse struct {
	Msg     string          `json:"msg"`
	Data    json.RawMessage `json:"data"`
	Code    int             `json:"code"`
	Success bool            `json:"success"`
}

type NoteListResponse struct {
	List      []Note `json:"list"`
	Total     int    `json:"total"`
	HasMore   bool   `json:"hasMore"`
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
}

type Note struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	NoteType    string   `json:"noteType"`
	Oneself     bool     `json:"oneself"`
	Link        string   `json:"link"`
	Description string   `json:"description"`
	Markdown    string   `json:"markdown"`
	ContainerID string   `json:"containerId"`
	TagIDs      []string `json:"tagIds"`
	AuthorID    string   `json:"authorId"`
	CreateTime  int64    `json:"createTime"`
	UpdateTime  int64    `json:"updateTime"`
}

func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) ListNotesByAuthor(authorID string, pageIndex, pageSize int) (*NoteListResponse, error) {
	url := fmt.Sprintf("%s/cli/note/list?authorId=%s&pageIndex=%d&pageSize=%d&sort=desc&sortBy=createTime",
		c.baseURL, authorID, pageIndex, pageSize)

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var result NoteListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("decode note list: %w", err)
	}
	return &result, nil
}

func (c *Client) GetNote(id string) (*Note, error) {
	url := fmt.Sprintf("%s/cli/note?id=%s", c.baseURL, id)

	body, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var note Note
	if err := json.Unmarshal(body, &note); err != nil {
		return nil, fmt.Errorf("decode note: %w", err)
	}
	return &note, nil
}

func (c *Client) doRequest(url string) (json.RawMessage, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var wrapper apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("decode wrapper: %w", err)
	}

	if wrapper.Msg != "" {
		return nil, fmt.Errorf("API error: %s", wrapper.Msg)
	}

	return wrapper.Data, nil
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("X-CLI-Token", c.token)
	req.Header.Set("x-client", "ReferenceAnswerRSS/1.0")
	req.Header.Set("Accept", "application/json")
}
