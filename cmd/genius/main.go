package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/joe-bricknell/genius"
)

// Song represents a song returned from the API
type Song struct {
	Title  string
	Artist string
}

//need to define our song struct
type Data struct {
	Response struct {
		Hits []hit
	} `json:"response"`
}

type hit struct {
	Highlights []interface{} `json:"highlights"`
	Index      string        `json:"index"`
	Type       string        `json:"type"`
	Result     struct {
		AnnotationCount          int    `json:"annotation_count"`
		APIPath                  string `json:"api_path"`
		FullTitle                string `json:"full_title"`
		HeaderImageThumbnailURL  string `json:"header_image_thumbnail_url"`
		HeaderImageURL           string `json:"header_image_url"`
		ID                       int    `json:"id"`
		LyricsOwnerID            int    `json:"lyrics_owner_id"`
		LyricsState              string `json:"lyrics_state"`
		Path                     string `json:"path"`
		PyongsCount              int    `json:"pyongs_count"`
		SongArtImageThumbnailURL string `json:"song_art_image_thumbnail_url"`
		SongArtImageURL          string `json:"song_art_image_url"`
		Stats                    struct {
			UnreviewedAnnotations int  `json:"unreviewed_annotations"`
			Concurrents           int  `json:"concurrents"`
			Hot                   bool `json:"hot"`
			Pageviews             int  `json:"pageviews"`
		} `json:"stats"`
		Title             string `json:"title"`
		TitleWithFeatured string `json:"title_with_featured"`
		URL               string `json:"url"`
		PrimaryArtist     struct {
			APIPath        string `json:"api_path"`
			HeaderImageURL string `json:"header_image_url"`
			ID             int    `json:"id"`
			ImageURL       string `json:"image_url"`
			IsMemeVerified bool   `json:"is_meme_verified"`
			IsVerified     bool   `json:"is_verified"`
			Name           string `json:"name"`
			URL            string `json:"url"`
			Iq             int    `json:"iq"`
		} `json:"primary_artist"`
	} `json:"result"`
}

type AllSongs struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Songs []struct {
			AnnotationCount          int    `json:"annotation_count"`
			APIPath                  string `json:"api_path"`
			FullTitle                string `json:"full_title"`
			HeaderImageThumbnailURL  string `json:"header_image_thumbnail_url"`
			HeaderImageURL           string `json:"header_image_url"`
			ID                       int    `json:"id"`
			LyricsOwnerID            int    `json:"lyrics_owner_id"`
			LyricsState              string `json:"lyrics_state"`
			Path                     string `json:"path"`
			PyongsCount              int    `json:"pyongs_count"`
			SongArtImageThumbnailURL string `json:"song_art_image_thumbnail_url"`
			SongArtImageURL          string `json:"song_art_image_url"`
			Stats                    struct {
				UnreviewedAnnotations int  `json:"unreviewed_annotations"`
				Concurrents           int  `json:"concurrents"`
				Hot                   bool `json:"hot"`
				Pageviews             int  `json:"pageviews"`
			} `json:"stats,omitempty"`
			Title             string `json:"title"`
			TitleWithFeatured string `json:"title_with_featured"`
			URL               string `json:"url"`
			PrimaryArtist     struct {
				APIPath        string `json:"api_path"`
				HeaderImageURL string `json:"header_image_url"`
				ID             int    `json:"id"`
				ImageURL       string `json:"image_url"`
				IsMemeVerified bool   `json:"is_meme_verified"`
				IsVerified     bool   `json:"is_verified"`
				Name           string `json:"name"`
				URL            string `json:"url"`
				Iq             int    `json:"iq"`
			} `json:"primary_artist,omitempty"`
		} `json:"songs"`
		NextPage int `json:"next_page"`
	} `json:"response"`
}

// Lyrics represents the lyrics returned from the lyric api
type Lyrics struct {
	Lyrics string `json:"lyrics"`
}

func main() {
	var searchFlag string
	var artistFlag string
	var wordFlag string
	flag.StringVar(&searchFlag, "search", "", "specify your search term")
	flag.StringVar(&artistFlag, "artist", "", "specify your search term")
	flag.StringVar(&wordFlag, "word", "", "specify the words you want to look for")
	flag.Parse()

	lyrics, err := genius.GetLyricsBySearch(searchFlag)
	if err != nil {
		panic(err)
	}

	lyrics, err = getAllLyricsByArtist(artistFlag)
	if err != nil {
		panic(err)
	}

	fmt.Println(lyrics)

	wordMap, err := findWords(lyrics, wordFlag)
	if err != nil {
		panic(err)
	}

	displayWordCount(wordMap)

}

// allSongsByArtist will return all the songs by a given artist
func getAllLyricsByArtist(artistFlag string) ([]Lyrics, error) {
	id, err := getArtistID(artistFlag)
	if err != nil {
		return nil, err
	}

	songs, err := songsByArtist(*id)

	lyrics, err := getLyrics(songs)
	if err != nil {
		return nil, err
	}

	return lyrics, nil
}

// getArtistID will call to the genius api search and pull out the artist id from the first search result
func getArtistID(artistFlag string) (*int, error) {
	encodedSearch := url.QueryEscape(artistFlag)
	// build request to genius api
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.genius.com/search?q=%v", encodedSearch), strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	token := "SWIZahaJ5gY3S8ZOAwLbTlpREdKOXMakvPPM_0vD5q1AXId4J4fGTDJ-VO-h0Ojp"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "Application/json")

	// make request to genius api
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// read the body of the request
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal json into song response struct
	var apiSongResponse data
	if err := json.Unmarshal(body, &apiSongResponse); err != nil {
		return nil, err
	}

	id := apiSongResponse.Response.Hits[0].Result.PrimaryArtist.ID

	return &id, nil
}

func songsByArtist(id int) ([]Song, error) {
	// build request using the id obtained earlier
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.genius.com/artists/%v/songs?sort=popularity", id), strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	token := "SWIZahaJ5gY3S8ZOAwLbTlpREdKOXMakvPPM_0vD5q1AXId4J4fGTDJ-VO-h0Ojp"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "Application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiSongResponse AllSongs
	if err := json.Unmarshal(body, &apiSongResponse); err != nil {
		return nil, err
	}

	songList, err := getSongs(apiSongResponse)
	if err != nil {
		return nil, err
	}

	return songList, nil
}

func getSongs(apiResponse AllSongs) ([]Song, error) {
	// define our song list variable and range over the songs and add the
	// song name and artist to the song struct
	var songList []Song
	for _, songs := range apiResponse.Response.Songs {
		song := Song{
			Title:  strings.TrimSpace(songs.Title),
			Artist: strings.TrimSpace(songs.PrimaryArtist.Name),
		}

		songList = append(songList, song)
	}

	return songList, nil
}

// searchSongs will call to the genius api and return a list of songs matching
// a particular search
func searchSongs(search string) ([]Song, error) {
	// build request to genius api
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.genius.com/search?q=%v", search), strings.NewReader(""))
	if err != nil {
		return nil, err
	}

	token := "SWIZahaJ5gY3S8ZOAwLbTlpREdKOXMakvPPM_0vD5q1AXId4J4fGTDJ-VO-h0Ojp"
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "Application/json")

	// make request to genius api
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	// read the body of the request
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// unmarshal json into song response struct
	var apiSongResponse data
	if err := json.Unmarshal(body, &apiSongResponse); err != nil {
		return nil, err
	}

	// define our song list variable and range over the songs and add the
	// song name and artist to the song struct
	var songList []Song
	for _, songs := range apiSongResponse.Response.Hits {
		song := Song{
			Title:  strings.TrimSpace(songs.Result.Title),
			Artist: strings.TrimSpace(songs.Result.PrimaryArtist.Name),
		}

		songList = append(songList, song)
	}

	return songList, nil
}

// getLyrics will call to the lyrics api and return the lyrics for a particular song
func getLyrics(songList []Song) ([]Lyrics, error) {
	var allLyrics []Lyrics
	var lyrics Lyrics
	for _, song := range songList[0:5] {
		fmt.Printf("Artist: %v, Song: %v\n\n", song.Artist, song.Title)

		//	build request to lyrics api
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.lyrics.ovh/v1/%v/%v", song.Artist, song.Title), strings.NewReader(""))
		if err != nil {
			return nil, err
		}

		// make request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}

		// read body of the response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// unmarshall json into lyrics struct
		if err := json.Unmarshal(body, &lyrics); err != nil {
			return nil, err
		}

		allLyrics = append(allLyrics, lyrics)
	}

	return allLyrics, nil
}

// findWords will search through the lyrics and count the number of matches
// for particular words
func findWords(allLyrics []Lyrics, wvar string) (map[string]int, error) {
	wordFlags := strings.Fields(wvar)
	fmt.Println("wordflags:", wordFlags)

	var lyricCount int
	var fuckCount int
	var shitCount int
	var bitchCount int
	var pussyCount int

	for _, lyrics := range allLyrics {
		for _, word := range strings.Fields(lyrics.Lyrics) {
			lyricCount++
			switch {
			case
				strings.Contains(strings.ToLower(word), "fuck"),
				strings.Contains(strings.ToLower(word), "f-ck"),
				strings.Contains(strings.ToLower(word), "f*ck"):
				fuckCount++
			case strings.Contains(strings.ToLower(word), "shit"):
				shitCount++
			case
				strings.Contains(strings.ToLower(word), "bitch"),
				strings.Contains(strings.ToLower(word), "b*tch"),
				strings.Contains(strings.ToLower(word), "b-tch"):
				bitchCount++
			case
				strings.Contains(strings.ToLower(word), "pussy"),
				strings.Contains(strings.ToLower(word), "p*ssy"),
				strings.Contains(strings.ToLower(word), "p-ssy"):
				pussyCount++
			}
		}
	}

	fmt.Printf("total words counted: %v\n", lyricCount)

	wordMap := map[string]int{
		"fuckCount":  fuckCount,
		"shitCount":  shitCount,
		"bitchCount": bitchCount,
		"pussy":      pussyCount,
	}

	return wordMap, nil
}

func displayWordCount(wordMap map[string]int) {
	// we range over the map to get the keys and store them in a slice
	keys := make([]string, 0, len(wordMap))
	for k := range wordMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	fmt.Printf("%v:%v,\n%v:%v,\n%v:%v,\n%v:%v\n",
		keys[0], wordMap[keys[0]],
		keys[1], wordMap[keys[1]],
		keys[2], wordMap[keys[2]],
		keys[3], wordMap[keys[3]])

}
