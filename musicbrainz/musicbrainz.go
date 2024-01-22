package musicbrainz

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Track struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Number   string `json:"number"`
	Position int    `json:"position"`

	//	          "number": "2",
	//	          "length": 324600,
	//	          "position": 2,
	//	          "title": "Sad but True",

	Recording struct {
		ArtistCredit []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
			//	                "joinphrase": "",
			//	                "name": "Metallica",
			//	                "artist": {
			//	                  "type": "Group",
			//	                  "sort-name": "Metallica",
			//	                  "type-id": "e431f5f6-b5d2-343d-8b36-72607fffb74b",
			//	                  "disambiguation": "",
			//	                  "name": "Metallica",
			//	                  "id": "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab"
			//	                }
		} `json:"artist-credit"`
		Disambiguation   string `json:"disambiguation"`
		Title            string `json:"title"`
		Length           int    `json:"length"`
		Id               string `json:"id"`
		FirstReleaseDate string `json:"first-release-date"`
		Video            bool   `json:"video"`
	} `json:"recording"`

	//	          "number": "2",
	//	          "length": 324600,
	//	          "position": 2,
	//	          "title": "Sad but True",
	//	          "artist-credit": [
	//	            {
	//	              "joinphrase": "",
	//	              "name": "Metallica",
	//	              "artist": {
	//	                "id": "65f4f0c5-ef9e-490c-aee3-909e7ae6b2ab",
	//	                "name": "Metallica",
	//	                "type-id": "e431f5f6-b5d2-343d-8b36-72607fffb74b",
	//	                "disambiguation": "",
	//	                "sort-name": "Metallica",
	//	                "type": "Group"
	//	              }
	//	            }
	//	          ]
}

type Media struct {
	Title    string `json:"title"`
	FormatId string `json:"format-id"`
	Position int    `json:"position"`
	Format   string `json:"format"`

	TrackCount  string `json:"track-count"`
	TrackOffset string `json:"track-offset"`

	Tracks []Track `json:"tracks"`
}

type Metadata struct {
	Id    string  `json:"id"`
	Title string  `json:"title"`
	Date  string  `json:"date"`
	Media []Media `json:"media"`

	ArtistCredit []struct {
		Name   string `json:"name"`
		Artist struct {
			Id   string `json:"id"`
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"arist"`
	} `json:"artist-credit"`

	//	{
	//	  "packaging-id": null,
	//	  "text-representation": {
	//	    "language": "eng",
	//	    "script": "Latn"
	//	  },
	//	  "title": "Metallica",
	//	  "cover-art-archive": {
	//	    "front": true,
	//	    "artwork": true,
	//	    "count": 6,
	//	    "back": true,
	//	    "darkened": false
	//	  },
	//	  "asin": "B000002H97",
	//	  "disambiguation": "",
	//	  "status": "Official",
	//	  "release-events": [
	//	    {
	//	      "date": "1991-08-12",
	//	      "area": {
	//	        "type-id": null,
	//	        "disambiguation": "",
	//	        "iso-3166-1-codes": [
	//	          "CA"
	//	        ],
	//	        "type": null,
	//	        "sort-name": "Canada",
	//	        "id": "71bbafaa-e825-3e15-8ca9-017dcad1748b",
	//	        "name": "Canada"
	//	      }
	//	    }
	//	  ],
	//	  "packaging": null,
	//	  "barcode": "075596111324",
	//	  "id": "2529f558-970b-33d2-a42c-41ab15a970c6",
	//	  "country": "CA",
	//	  "quality": "high",
	//	  "date": "1991-08-12",
	//	  "artist-credit": [
	//	    {
	//	    }
	//	  ],
	//	  "status-id": "4e304316-386d-3409-af2e-78857eec5cfe"
	//	}
}

type CoverArtResponse struct {
	Ext  string
	Data []byte
}

func FetchCoverArt(mbid string) (CoverArtResponse, error) {
	// https://coverartarchive.org/release/{mbid}/front
	url := fmt.Sprintf("https://coverartarchive.org/release/%v/front", mbid)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return CoverArtResponse{}, err
	}

	req.Header.Set("User-Agent", "dwebble/0.0.1 ( github.com/nanoteck137/dwebble )")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return CoverArtResponse{}, err
	}
	defer res.Body.Close()

	contentType := res.Header.Get("Content-Type")
	var ext string

	switch contentType {
	case "image/jpeg":
		ext = "jpg"
	case "image/png":
		ext = "png"
	default:
		return CoverArtResponse{}, fmt.Errorf("Unknown content type for cover art: %v\n", contentType)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return CoverArtResponse{}, err
	}

	return CoverArtResponse{
		Ext:  ext,
		Data: data,
	}, err
}

func FetchAlbumMetadata(mbid string) (Metadata, error) {
	// https://musicbrainz.org/ws/2/release/{mbid}?inc=artist-credits%2Brecordings&fmt=json
	url := fmt.Sprintf("https://musicbrainz.org/ws/2/release/%v?inc=artist-credits+recordings&fmt=json", mbid)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Metadata{}, err
	}

	req.Header.Set("User-Agent", "dwebble/0.0.1 ( github.com/nanoteck137/dwebble )")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return Metadata{}, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return Metadata{}, err
	}

	var metadata Metadata
	json.Unmarshal(data, &metadata)

	return metadata, nil
}

func (metadata *Metadata) DebugDump() {
	fmt.Printf("Title: %v\n", metadata.Title)
	fmt.Printf("Date: %v\n", metadata.Date)
	for _, m := range metadata.Media {
		fmt.Printf("  %v\n", m.Position)

		for _, t := range m.Tracks {
			fmt.Printf("    %v", t.Recording.Title)
			fmt.Printf(" - %v\n", t.Recording.ArtistCredit[0].Name)
		}
	}
}
