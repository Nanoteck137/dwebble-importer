package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"

	"github.com/kr/pretty"
	"github.com/nanoteck137/dwebble-importer/musicbrainz"
	"github.com/nanoteck137/dwebble-importer/utils"
	"github.com/nanoteck137/dwebble/types"
)

func CreateArtist(name string) (types.ApiPostArtistData, error) {
	var b bytes.Buffer
	form := multipart.NewWriter(&b)

	err := form.WriteField("name", name)
	if err != nil {
		return types.ApiPostArtistData{}, err
	}

	err = form.Close()
	if err != nil {
		return types.ApiPostArtistData{}, err
	}

	req, err := http.NewRequest("POST", "http://localhost:3000/api/v1/artists", &b)
	if err != nil {
		return types.ApiPostArtistData{}, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.ApiPostArtistData{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return types.ApiPostArtistData{}, err
	}

	if res.StatusCode != 200 {
		return types.ApiPostArtistData{}, errors.New("Request error: " + string(data))
	}

	var response types.ApiResponse[types.ApiPostArtistData]
	err = json.Unmarshal(data, &response)
	if err != nil {
		return types.ApiPostArtistData{}, err
	}

	pretty.Println(response)

	return response.Data, nil
}

type Config struct {
	Typ    string `toml:"type"`
	Name   string `toml:"name"`
	Artist string `toml:"artist"`

	Tracks []struct {
		Num      int    `toml:"num"`
		Name     string `toml:"name"`
		Filename string `toml:"filename"`
		Artist   string `toml:"artist"`
	} `toml:"tracks"`
}

func main() {
	// artist, err := CreateArtist("test")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// pretty.Println(artist)

	// source := "/Volumes/media/music/Ado/unravel"
	source := "/Volumes/media/musicraw/Metallica/Metallica"

	// mbid := "a03d4434-6b7c-4d56-adb7-f08d478fc09f"
	mbid := "69a8ca83-a182-3375-a702-a30e216748c9"

	dirEntries, err := os.ReadDir(source)
	if err != nil {
		log.Fatal(err)
	}

	var validTracks []utils.FileResult

	for _, entry := range dirEntries {
		p := path.Join(source, entry.Name())

		ext := path.Ext(p)[1:]
		if utils.IsValidTrackExt(ext) {
			res, err := utils.CheckFile(p)
			if err != nil {
				log.Fatal(err)
			}

			validTracks = append(validTracks, res)
		}

		fmt.Printf("%v\n", p)
	}

	pretty.Println(validTracks)

	metadata, err := musicbrainz.FetchAlbumMetadata(mbid)
	if err != nil {
		log.Fatal(err)
	}

	if len(metadata.Media) > 1 {
		log.Fatal("Only support single disc media")
	}

	media := metadata.Media[0]

	if len(media.Tracks) != len(validTracks) {
		log.Fatal("Track length mismatch")
	}

	type UnprocessedTrack struct {
		name     string
		number   int
		filePath string
	}

	var unprocessedTracks []UnprocessedTrack

	for _, track := range validTracks {
		var metadataTrack musicbrainz.Track
		found := false
		for _, mt := range media.Tracks {
			if track.Number == mt.Position {
				metadataTrack = mt
				found = true
			}
		}

		if found {
			fmt.Printf("%v -> %v\n", track.Path, metadataTrack.Title)
			unprocessedTracks = append(unprocessedTracks, UnprocessedTrack{
				name:     metadataTrack.Title,
				number:   track.Number,
				filePath: track.Path,
			})
		}
	}

	pretty.Println(unprocessedTracks)

	for _, track := range unprocessedTracks {
		_ = track
		// TODO(patrik): Create Best Quality File (flac)
		// TODO(patrik): Create Mobile Quality File (flac)
	}

	// if len(metadata.

	// pretty.Println(metadata)

	// type Album struct {
	// }

	type Track struct {
		name   string
		number int

		bestQualityFilePath   string
		mobileQualityFilePath string
		coverArtFilePath      string

		albumId  string
		artistId string
	}

	// os.MkdirTemp()
}
