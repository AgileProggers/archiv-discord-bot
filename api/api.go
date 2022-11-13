package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Vod struct {
	UUID           string     `json:"uuid"`
	Title          string     `json:"title"`
	Date           time.Time  `json:"date"`
	Filename       string     `json:"filename"`
	Viewcount      int        `json:"view_count"`
	Clips          []struct{} `json:"clips"`
	TitleRank      float64    `json:"title_rank"`
	TranscriptRank float64    `json:"transcript_rank"`
}

type SearchResponse struct {
	Error  bool  `json:"error"`
	Result []Vod `json:"result"`
}

type UUIDResponse struct {
	Error  bool `json:"error"`
	Result Vod  `json:"result"`
}

type StatsResponse struct {
	Error  bool `json:"error"`
	Result struct {
		CountVodsTotal       int     `json:"count_vods_total"`
		CountClipsTotal      int     `json:"count_clips_total"`
		CountHStreamed       float64 `json:"count_h_streamed"`
		CountSizeBytes       int     `json:"count_size_bytes"`
		CountTranscriptWords int     `json:"count_transcript_words"`
		CountUniqueWords     int     `json:"count_unique_words"`
		CountAvgWords        int     `json:"count_avg_words"`
		DatabaseSize         int     `json:"database_size"`
		ClipsPerCreator      []struct {
			Name      string `json:"name"`
			ClipCount int    `json:"clip_count"`
			ViewCount int    `json:"view_Count"`
		} `json:"clips_per_creator"`
	} `json:"result"`
}

func Search(response *SearchResponse, query string, limit int) error {
	var requestURL string
	if limit == 1 {
		requestURL = fmt.Sprintf("https://api.wubbl0rz.tv/vods/?limit=%d", limit)
	} else {
		requestURL = fmt.Sprintf("https://api.wubbl0rz.tv/vods/?limit=%d&q=%s", limit, query)
	}
	res, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("status code was %d, expected 200", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	return nil
}

func UUID(response *UUIDResponse, uuid string) error {
	requestURL := fmt.Sprintf("https://api.wubbl0rz.tv/vods/%s", uuid)
	res, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("status code was %d, expected 200", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	return nil
}

func Stats(response *StatsResponse) error {
	res, err := http.Get("https://api.wubbl0rz.tv/stats/long")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("status code was %d, expected 200", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return err
	}

	return nil
}
