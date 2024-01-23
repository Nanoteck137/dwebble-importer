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
	"net/textproto"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/kr/pretty"
	"github.com/nanoteck137/dwebble-importer/musicbrainz"
	"github.com/nanoteck137/dwebble-importer/utils"
	"github.com/nanoteck137/dwebble/types"
)

type Track struct {
	name   string
	number int

	bestQualityFilePath   string
	mobileQualityFilePath string
	coverArtFilePath      string

	albumId  string
	artistId string
}

func GetArtists(name string) (types.ApiGetArtistsData, error) {
	url := fmt.Sprintf("http://localhost:3000/api/v1/artists?name=%v", url.QueryEscape(name))
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return types.ApiGetArtistsData{}, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.ApiGetArtistsData{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return types.ApiGetArtistsData{}, err
	}

	if res.StatusCode != 200 {
		return types.ApiGetArtistsData{}, errors.New("Request error: " + string(data))
	}

	var response types.ApiResponse[types.ApiGetArtistsData]
	err = json.Unmarshal(data, &response)
	if err != nil {
		return types.ApiGetArtistsData{}, err
	}

	pretty.Println(response)

	return response.Data, nil

}

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

func CreateAlbum(name, artistId string) (types.ApiPostAlbumData, error) {
	var b bytes.Buffer
	form := multipart.NewWriter(&b)

	err := form.WriteField("name", name)
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}

	err = form.WriteField("artist", artistId)
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}

	// form.CreateFormFile()

	err = form.Close()
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}

	req, err := http.NewRequest("POST", "http://localhost:3000/api/v1/albums", &b)
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}

	if res.StatusCode != 200 {
		return types.ApiPostAlbumData{}, errors.New("Request error: " + string(data))
	}

	var response types.ApiResponse[types.ApiPostAlbumData]
	err = json.Unmarshal(data, &response)
	if err != nil {
		return types.ApiPostAlbumData{}, err
	}

	pretty.Println(response)

	return response.Data, nil
}

func CreateTrack(track Track) (types.ApiPostTrackData, error) {
	var b bytes.Buffer
	form := multipart.NewWriter(&b)

	err := form.WriteField("name", track.name)
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	err = form.WriteField("number", strconv.Itoa(track.number))
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	err = form.WriteField("album", track.albumId)
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	err = form.WriteField("artist", track.artistId)
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	createFileField := func(fieldName, filePath, contentType string) error {
		fileName := path.Base(filePath)

		h := make(textproto.MIMEHeader)
		dis := fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileName)
		h.Set("Content-Disposition", dis)
		h.Set("Content-Type", contentType)

		writer, err :=  form.CreatePart(h)
		if err != nil {
			return err
		}

		fileData, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		_, err = writer.Write(fileData)
		return err
	}

	createFileField("bestQualityFile", track.bestQualityFilePath, "audio/flac")
	createFileField("mobileQualityFile", track.mobileQualityFilePath, "audio/mpeg")
	createFileField("coverArt", track.coverArtFilePath, "image/png")

	// createFileField()

	// form.CreateFormFile()

	err = form.Close()
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	req, err := http.NewRequest("POST", "http://localhost:3000/api/v1/tracks", &b)
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return types.ApiPostTrackData{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return types.ApiPostTrackData{}, err
	}

	if res.StatusCode != 200 {
		return types.ApiPostTrackData{}, errors.New("Request error: " + string(data))
	}

	var response types.ApiResponse[types.ApiPostTrackData]
	err = json.Unmarshal(data, &response)
	if err != nil {
		return types.ApiPostTrackData{}, err
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
	// source := "/mnt/media/musicraw/Metallica/Metallica"

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

	artists := make(map[string]string)

	pretty.Println(metadata)
	metadataArtist := metadata.ArtistCredit[0]
	artists[metadataArtist.Name] = ""

	// album, err := CreateAlbum(metadata.Title, artist.Id)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// _ = album

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
		artist   string
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
			artistName := metadataTrack.Recording.ArtistCredit[0].Name
			artists[artistName] = ""
			unprocessedTracks = append(unprocessedTracks, UnprocessedTrack{
				name:     metadataTrack.Title,
				number:   track.Number,
				filePath: track.Path,
				artist:   artistName,
			})
		}
	}

	for k := range artists {
		res, err := GetArtists(fmt.Sprintf("%%%v%%", k))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Res: %v\n", res)

		if len(res.Artists) == 0 {
			artist, err := CreateArtist(k)
			if err != nil {
				log.Fatal(err)
			}
			artists[k] = artist.Id
		} else {
			if len(res.Artists) > 1 {
				log.Printf("Need to handle when the server returns multiple artists")
			}

			artist := res.Artists[0]
			artists[k] = artist.Id
		}
	}

	pretty.Println(unprocessedTracks)

	// dir, err := os.MkdirTemp(".", "testing")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer os.RemoveAll(dir)

	dir := "./testing"
	err = os.MkdirAll(dir, 0766)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Dir: %v\n", dir)

	artistName := metadata.ArtistCredit[0].Name

	album, err := CreateAlbum(metadata.Title, artists[artistName])
	if err != nil {
		log.Fatal(err)
	}

	var processedTracks []Track

	for _, track := range unprocessedTracks {
		_ = track
		// TODO(patrik): Create Best Quality File (flac)
		// TODO(patrik): Create Mobile Quality File (flac)

		// TODO(patrik): Check extention
		dstName := fmt.Sprintf("%v.best.flac", track.number)
		bestQualityFilePath := path.Join(dir, dstName)
		utils.RunFFmpeg(true, "-i", track.filePath, bestQualityFilePath)

		// ffmpeg -i input.flac -ab 320k -map_metadata 0 -id3v2_version 3 output.mp3
		dstName = fmt.Sprintf("%v.mobile.mp3", track.number)
		mobileQualityFile := path.Join(dir, dstName)
		utils.RunFFmpeg(true, "-i", track.filePath, "-ab", "192k", mobileQualityFile)

		processedTracks = append(processedTracks, Track{
			name:                  track.name,
			number:                track.number,
			bestQualityFilePath:   bestQualityFilePath,
			mobileQualityFilePath: mobileQualityFile,
			coverArtFilePath:      "../test.png",
			albumId:               album.Id,
			artistId:              artists[track.artist],
		})
	}

	for _, track := range processedTracks {
		res, err := CreateTrack(track)
		if err != nil {
			log.Fatal(err)
		}

		pretty.Println(res)
	}

	// pretty.Println(processedTracks)
	// pretty.Println(artists)
}
