package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/kr/pretty"
	"github.com/nanoteck137/dwebble-importer/server"
	"github.com/nanoteck137/dwebble-importer/utils"
	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/cobra"
)

type UnprocessedTrack struct {
	Name      string
	Number    int
	AlbumId   string
	ArtistId  string
	TrackFile string
	// BestQualityFile   string
	// MobileQualityFile string
	// CoverArt          string
}

type ProcessedTrack struct {
	Name              string
	Number            int
	AlbumId           string
	ArtistId          string
	BestQualityFile   string
	MobileQualityFile string
	CoverArt          string
}

type Track struct {
	name   string
	number int

	bestQualityFilePath   string
	mobileQualityFilePath string
	coverArtFilePath      string

	albumId  string
	artistId string
}

type ConfigTrack struct {
	Num      int    `toml:"num"`
	Name     string `toml:"name"`
	Filename string `toml:"filename"`
	Artist   string `toml:"artist"`
}

type Config struct {
	Typ    string `toml:"type"`
	Name   string `toml:"name"`
	Artist string `toml:"artist"`

	Tracks []ConfigTrack `toml:"tracks"`
}

var rootCmd = &cobra.Command{
	Use:     "dwebble-import",
	Version: "v0.0.1",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Run")
	},
}

var createConfigCmd = &cobra.Command{
	Use:   "create-config",
	Short: "Create new album config",
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := "./"
		if len(args) > 0 {
			dir = args[0]
		}
		runCreateConfig(dir)
	},
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import album to dwebble server",
	Run: func(cmd *cobra.Command, args []string) {
		serverAddr, _ := cmd.Flags().GetString("serverAddr")
		fmt.Printf("Import %v\n", serverAddr)
	},
}

func init() {
	importCmd.PersistentFlags().StringP("serverAddr", "s", "", "")

	rootCmd.AddCommand(createConfigCmd)
	rootCmd.AddCommand(importCmd)
}

func runCreateConfig(dir string) {
	fmt.Printf("Dir: %v\n", dir)

	entries, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var fileResults []utils.FileResult

	for _, entry := range entries {
		p := path.Join(dir, entry.Name())

		ext := path.Ext(p)[1:]
		if utils.IsValidTrackExt(ext) {
			res, err := utils.CheckFile(p)
			if err != nil {
				log.Fatal(err)
			}

			fileResults = append(fileResults, res)
		}
	}

	albumArtistName := ""
	albumName := ""
	var tracks []ConfigTrack

	for _, file := range fileResults {
		if file.Probe.Track != -1 && file.Probe.Track != file.Number {
			log.Fatal("Track number not matching")
		}

		if file.Probe.AlbumArtist != "" {
			albumArtistName = file.Probe.AlbumArtist
		}

		if file.Probe.Album != "" {
			albumName = file.Probe.Album
		}

		tracks = append(tracks, ConfigTrack{
			Num:      file.Number,
			Name:     file.Probe.Title,
			Filename: path.Base(file.Path),
			Artist:   file.Probe.Artist,
		})
	}

	sort.SliceStable(tracks, func(i, j int) bool {
		return tracks[i].Num < tracks[j].Num
	})

	config := Config{
		Typ:    "",
		Name:   albumName,
		Artist: albumArtistName,
		Tracks: tracks,
	}

	data, err := toml.Marshal(config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(string(data))

	configPath := path.Join(dir, "album.toml")
	if _, err := os.Stat(configPath); !errors.Is(err, os.ErrNotExist) {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Config already exists overwrite (y/n): ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch text {
		case "y", "yes":
			fmt.Printf("Writing config\n")
			os.WriteFile(configPath, data, 0644)
		default:
			fmt.Printf("Not writing config\n")
		}
	} else {
		os.WriteFile(configPath, data, 0644)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}

	return
	api := server.New("http://localhost:3000/api/v1")

	// d := "/Volumes/media/musicraw/Metallica/Metallica"
	d := "/Volumes/media/music/Ado/unravel"

	p := path.Join(d, "album.toml")
	data, err := os.ReadFile(p)
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal(err)
	}

	pretty.Println(config)

	allArtists := make(map[string]string)

	allArtists[config.Name] = ""

	for _, track := range config.Tracks {
		if track.Artist != "" {
			allArtists[track.Artist] = ""
		}
	}

	for name := range allArtists {
		res, err := api.GetArtists(name)
		if err != nil {
			log.Fatal(err)
		}

		if len(res.Artists) == 0 {
			artist, err := api.CreateArtist(server.ArtistData{
				Name:    name,
				Picture: nil,
			})

			if err != nil {
				log.Fatal(err)
			}

			allArtists[name] = artist.Id
		} else {
			if len(res.Artists) > 1 {
				log.Fatalf("Server returned more then one artist for name '%s'", name)
			}

			allArtists[name] = res.Artists[0].Id
		}
	}

	artistId := allArtists[config.Artist]
	albums, err := api.GetArtistAlbums(artistId, config.Name)
	if err != nil {
		log.Fatal(err)
	}

	albumId := ""

	if len(albums.Albums) == 0 {
		album, err := api.CreateAlbum(server.AlbumData{
			Name:     config.Name,
			ArtistId: artistId,
			CoverArt: nil,
		})

		if err != nil {
			log.Fatal(err)
		}

		pretty.Println(album)
		albumId = album.Id
	} else {
		if len(albums.Albums) > 1 {
			log.Fatalf("Server returned more then one album for '%v' - '%v'", config.Artist, config.Name)
		}

		albumId = albums.Albums[0].Id
	}

	var unprocessedTracks []UnprocessedTrack

	for _, track := range config.Tracks {
		artist := config.Artist
		if track.Artist != "" {
			artist = track.Artist
		}

		trackFile := path.Join(d, track.Filename)

		unprocessedTracks = append(unprocessedTracks, UnprocessedTrack{
			Name:      track.Name,
			Number:    track.Num,
			AlbumId:   albumId,
			ArtistId:  allArtists[artist],
			TrackFile: trackFile,
		})
	}

	pretty.Println(unprocessedTracks)

	dir := "./testing"
	err = os.MkdirAll(dir, 0766)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Dir: %v\n", dir)

	var processedTracks []ProcessedTrack

	for _, track := range unprocessedTracks {
		// TODO(patrik): Check extention
		dstName := fmt.Sprintf("%v.best.flac", track.Number)
		bestQualityFilePath := path.Join(dir, dstName)
		utils.RunFFmpeg(true, "-i", track.TrackFile, "-map_metadata", "-1", "-map", "0", "-map", "-0:v", "-c:a", "copy", bestQualityFilePath)

		// ffmpeg -i input.flac -ab 320k -map_metadata 0 -id3v2_version 3 output.mp3
		dstName = fmt.Sprintf("%v.mobile.mp3", track.Number)
		mobileQualityFile := path.Join(dir, dstName)
		utils.RunFFmpeg(true, "-i", track.TrackFile, "-ab", "192k", mobileQualityFile)

		processedTracks = append(processedTracks, ProcessedTrack{
			Name:              track.Name,
			Number:            track.Number,
			AlbumId:           track.AlbumId,
			ArtistId:          track.ArtistId,
			BestQualityFile:   bestQualityFilePath,
			MobileQualityFile: mobileQualityFile,
			CoverArt:          "./testing/test.png",
		})
	}

	getContentTypeFromExt := func(ext string) string {
		switch ext {
		case "flac":
			return "audio/flac"
		case "mp3":
			return "audio/mpeg"
		case "png":
			return "image/png"
		default:
			log.Fatalf("Unsupported ext '%v'", ext)
		}

		return ""
	}

	createFile := func(filePath string) (server.File, error) {
		ext := path.Ext(filePath)[1:]
		contentType := getContentTypeFromExt(ext)
		name := path.Base(filePath)

		content, err := os.Open(filePath)
		if err != nil {
			return server.File{}, err
		}

		return server.File{
			ContentType: contentType,
			Name:        name,
			Content:     content,
		}, nil
	}

	for _, track := range processedTracks {
		bestQualityFile, err := createFile(track.BestQualityFile)
		if err != nil {
			log.Fatal(err)
		}

		mobileQualityFile, err := createFile(track.MobileQualityFile)
		if err != nil {
			log.Fatal(err)
		}

		coverArt, err := createFile(track.CoverArt)
		if err != nil {
			log.Fatal(err)
		}

		_, err = api.CreateTrack(server.TrackData{
			Name:              track.Name,
			Number:            track.Number,
			AlbumId:           track.AlbumId,
			ArtistId:          track.ArtistId,
			BestQualityFile:   bestQualityFile,
			MobileQualityFile: mobileQualityFile,
			CoverArt:          coverArt,
		})

		if err != nil {
			log.Println(err)
		}
	}

	pretty.Println(processedTracks)

	{

		// unprocessedTracks := []UnprocessedTrack{
		// 	{
		// 		Name:      "Enter Sandman",
		// 		Number:    1,
		// 		AlbumId:   album.Id,
		// 		ArtistId:  "",
		// 		TrackFile: "/Volumes/media/musicraw/Metallica/Metallica/track01.cdda.wav",
		// 	},
		// }
		//
		// _ = unprocessedTracks
		//
		// processedTracks := []ProcessedTrack{
		// 	{
		// 		Name:              "Test Track",
		// 		Number:            1,
		// 		AlbumId:           album.Id,
		// 		ArtistId:          artistId,
		// 		BestQualityFile:   "./testing/1.best.flac",
		// 		MobileQualityFile: "./testing/1.mobile.mp3",
		// 		CoverArt:          "./testing/test.png",
		// 	},
		// }
		//
		// getContentTypeFromExt := func(ext string) string {
		// 	switch ext {
		// 	case "flac":
		// 		return "audio/flac"
		// 	case "mp3":
		// 		return "audio/mpeg"
		// 	case "png":
		// 		return "image/png"
		// 	default:
		// 		log.Fatalf("Unsupported ext '%v'", ext)
		// 	}
		//
		// 	return ""
		// }
		//
		// createFile := func(filePath string) (server.File, error) {
		// 	ext := path.Ext(filePath)[1:]
		// 	contentType := getContentTypeFromExt(ext)
		// 	name := path.Base(filePath)
		//
		// 	content, err := os.Open(filePath)
		// 	if err != nil {
		// 		return server.File{}, err
		// 	}
		//
		// 	return server.File{
		// 		ContentType: contentType,
		// 		Name:        name,
		// 		Content:     content,
		// 	}, nil
		// }
		//
		// for _, track := range processedTracks {
		// 	bestQualityFile, err := createFile(track.BestQualityFile)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		//
		// 	mobileQualityFile, err := createFile(track.MobileQualityFile)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		//
		// 	coverArt, err := createFile(track.CoverArt)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		//
		// 	track, err := api.CreateTrack(server.TrackData{
		// 		Name:              track.Name,
		// 		Number:            track.Number,
		// 		AlbumId:           track.AlbumId,
		// 		ArtistId:          track.ArtistId,
		// 		BestQualityFile:   bestQualityFile,
		// 		MobileQualityFile: mobileQualityFile,
		// 		CoverArt:          coverArt,
		// 	})
		//
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		//
		// 	pretty.Println(track)
		// }
	}

	return

	// artist, err := CreateArtist("test")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// pretty.Println(artist)

	// // source := "/Volumes/media/music/Ado/unravel"
	// // source := "/Volumes/media/musicraw/Metallica/Metallica"
	// source := "/mnt/media/musicraw/Metallica/Metallica"
	//
	// // mbid := "a03d4434-6b7c-4d56-adb7-f08d478fc09f"
	// mbid := "69a8ca83-a182-3375-a702-a30e216748c9"
	//
	// dirEntries, err := os.ReadDir(source)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// var validTracks []utils.FileResult
	//
	// for _, entry := range dirEntries {
	// 	p := path.Join(source, entry.Name())
	//
	// 	ext := path.Ext(p)[1:]
	// 	if utils.IsValidTrackExt(ext) {
	// 		res, err := utils.CheckFile(p)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	//
	// 		validTracks = append(validTracks, res)
	// 	}
	//
	// 	fmt.Printf("%v\n", p)
	// }
	//
	// pretty.Println(validTracks)
	//
	// metadata, err := musicbrainz.FetchAlbumMetadata(mbid)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// artists := make(map[string]string)
	//
	// pretty.Println(metadata)
	// metadataArtist := metadata.ArtistCredit[0]
	// artists[metadataArtist.Name] = ""
	//
	// // album, err := CreateAlbum(metadata.Title, artist.Id)
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// //
	// // _ = album
	//
	// if len(metadata.Media) > 1 {
	// 	log.Fatal("Only support single disc media")
	// }
	//
	// media := metadata.Media[0]
	//
	// if len(media.Tracks) != len(validTracks) {
	// 	log.Fatal("Track length mismatch")
	// }
	//
	// type UnprocessedTrack struct {
	// 	name     string
	// 	number   int
	// 	filePath string
	// 	artist   string
	// }
	//
	// var unprocessedTracks []UnprocessedTrack
	//
	// for _, track := range validTracks {
	// 	var metadataTrack musicbrainz.Track
	// 	found := false
	// 	for _, mt := range media.Tracks {
	// 		if track.Number == mt.Position {
	// 			metadataTrack = mt
	// 			found = true
	// 		}
	// 	}
	//
	// 	if found {
	// 		fmt.Printf("%v -> %v\n", track.Path, metadataTrack.Title)
	// 		artistName := metadataTrack.Recording.ArtistCredit[0].Name
	// 		artists[artistName] = ""
	// 		unprocessedTracks = append(unprocessedTracks, UnprocessedTrack{
	// 			name:     metadataTrack.Title,
	// 			number:   track.Number,
	// 			filePath: track.Path,
	// 			artist:   artistName,
	// 		})
	// 	}
	// }
	//
	// for k := range artists {
	// 	res, err := GetArtists(fmt.Sprintf("%%%v%%", k))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	fmt.Printf("Res: %v\n", res)
	//
	// 	if len(res.Artists) == 0 {
	// 		artist, err := CreateArtist(k)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		artists[k] = artist.Id
	// 	} else {
	// 		if len(res.Artists) > 1 {
	// 			log.Printf("Need to handle when the server returns multiple artists")
	// 		}
	//
	// 		artist := res.Artists[0]
	// 		artists[k] = artist.Id
	// 	}
	// }
	//
	// pretty.Println(unprocessedTracks)
	//
	// // dir, err := os.MkdirTemp(".", "testing")
	// // if err != nil {
	// // 	log.Fatal(err)
	// // }
	// // defer os.RemoveAll(dir)
	//
	// dir := "./testing"
	// err = os.MkdirAll(dir, 0766)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// fmt.Printf("Dir: %v\n", dir)
	//
	// artistName := metadata.ArtistCredit[0].Name
	//
	// album, err := CreateAlbum(metadata.Title, artists[artistName])
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//
	// var processedTracks []Track
	//
	// for _, track := range unprocessedTracks {
	// 	_ = track
	// 	// TODO(patrik): Create Best Quality File (flac)
	// 	// TODO(patrik): Create Mobile Quality File (flac)
	//
	// 	// TODO(patrik): Check extention
	// 	dstName := fmt.Sprintf("%v.best.flac", track.number)
	// 	bestQualityFilePath := path.Join(dir, dstName)
	// 	utils.RunFFmpeg(true, "-i", track.filePath, bestQualityFilePath)
	//
	// 	// ffmpeg -i input.flac -ab 320k -map_metadata 0 -id3v2_version 3 output.mp3
	// 	dstName = fmt.Sprintf("%v.mobile.mp3", track.number)
	// 	mobileQualityFile := path.Join(dir, dstName)
	// 	utils.RunFFmpeg(true, "-i", track.filePath, "-ab", "192k", mobileQualityFile)
	//
	// 	processedTracks = append(processedTracks, Track{
	// 		name:                  track.name,
	// 		number:                track.number,
	// 		bestQualityFilePath:   bestQualityFilePath,
	// 		mobileQualityFilePath: mobileQualityFile,
	// 		coverArtFilePath:      "../test.png",
	// 		albumId:               album.Id,
	// 		artistId:              artists[track.artist],
	// 	})
	// }
	//
	// for _, track := range processedTracks {
	// 	res, err := CreateTrack(track)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	//
	// 	pretty.Println(res)
	// }

	// pretty.Println(processedTracks)
	// pretty.Println(artists)
}
