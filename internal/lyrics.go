package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/joe-bricknell/genius/internal/log"

	"github.com/joe-bricknell/genius/internal/models"
)

type LyricsResponse struct {
	Lyrics string `json:"lyrics"`
}

// getLyrics will call to the lyrics api and return the lyrics for a particular Song
func getLyrics(songList []models.Song) ([]models.Lyrics, error) {
	// create error channel to receive errors from go routines
	errCh := make(chan error)
	resultCh := make(chan models.Lyrics)

	allLyrics := make([]models.Lyrics, 0, 20)

	var wg sync.WaitGroup

	wg.Add(len(songList))

	log.Logger.Info("looking for lyrics for...")
	for _, song := range songList {
		log.Logger.Infof("%v - %v", song.Artist, song.Title)
		go doRequests(resultCh, errCh, &wg, song)
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(resultCh)
	}()

	go func() {
		for lyrics := range resultCh {
			allLyrics = append(allLyrics, lyrics)
		}
	}()

	for err := range errCh {
		return nil, err
	}

	return allLyrics, nil
}

func doRequests(resultCh chan<- models.Lyrics, errCh chan<- error, wg *sync.WaitGroup, song models.Song) {
	if wg != nil {
		defer wg.Done()
	}

	var lyricsResp LyricsResponse
	endpoint := fmt.Sprintf("%v/%v", song.Artist, song.Title)

	resp, err := makeRequestLyrics(endpoint)
	if err != nil {
		errCh <- err
		return
	}

	defer resp.Body.Close()

	// read body of the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err := fmt.Errorf("failed to read response body: %w", err)
		log.Logger.Errorf("doRequests failed: %v", err)

		errCh <- err
		return
	}

	// unmarshal json into lyrics struct
	if err := json.Unmarshal(body, &lyricsResp); err != nil {
		err := fmt.Errorf("failed to unmarshal response body: %w", err)
		log.Logger.Errorf("doRequests failed: %v", err)

		errCh <- err
		return
	}

	if lyricsResp.Lyrics == "" {
		log.Logger.Infof("failed to find lyrics for: %v - %v", song.Artist, song.Title)
	}

	lyrics := models.Lyrics{
		ID:     song.ID,
		Lyrics: lyricsResp.Lyrics,
	}

	resultCh <- lyrics
}
