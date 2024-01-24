package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strconv"

	"github.com/nanoteck137/dwebble/types"
)

type Server struct {
	baseUrl string
}

func New(baseUrl string) *Server {
	return &Server{
		baseUrl: baseUrl,
	}
}

type ArtistData struct {
	Name    string
	Picture io.Reader
}

func (server *Server) newReq(method, endpoint string, body io.Reader) (*http.Request, error) {
	url := server.baseUrl + endpoint
	return http.NewRequest("POST", url, body)
}

func (server *Server) CreateArtist(data ArtistData) (*types.ApiPostArtistData, error) {
	var b bytes.Buffer
	form := multipart.NewWriter(&b)

	if err := form.WriteField("name", data.Name); err != nil {
		return nil, err
	}

	if err := form.Close(); err != nil {
		return nil, err
	}

	req, err := server.newReq("POST", "/artists", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("Request error: " + string(body))
	}

	var response types.ApiResponse[types.ApiPostArtistData]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

type AlbumData struct {
	Name     string
	ArtistId string
	CoverArt io.Reader
}

func (server *Server) CreateAlbum(data AlbumData) (*types.ApiPostAlbumData, error) {
	var b bytes.Buffer
	form := multipart.NewWriter(&b)

	if err := form.WriteField("name", data.Name); err != nil {
		return nil, err
	}

	if err := form.WriteField("artist", data.ArtistId); err != nil {
		return nil, err
	}

	if err := form.Close(); err != nil {
		return nil, err
	}

	req, err := server.newReq("POST", "/albums", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("Request error: " + string(body))
	}

	var response types.ApiResponse[types.ApiPostAlbumData]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}

type File struct {
	ContentType string
	Name        string
	Content     io.Reader
}

type TrackData struct {
	Name              string
	Number            int
	AlbumId           string
	ArtistId          string
	BestQualityFile   File
	MobileQualityFile File
	CoverArt          File
}

func createFileField(form *multipart.Writer, fieldName string, file *File) error {
	h := make(textproto.MIMEHeader)
	dis := fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, file.Name)
	h.Set("Content-Disposition", dis)
	h.Set("Content-Type", file.ContentType)

	writer, err := form.CreatePart(h)
	if err != nil {
		return err
	}

	data, err := io.ReadAll(file.Content)
	if err != nil {
		return err
	}

	fmt.Printf("File Size: %v -> %v\n", fieldName, len(data))

	_, err = writer.Write(data)
	return err
}

func (server *Server) CreateTrack(data TrackData) (*types.ApiPostTrackData, error) {
	var b bytes.Buffer
	form := multipart.NewWriter(&b)

	if err := form.WriteField("name", data.Name); err != nil {
		return nil, err
	}

	if err := form.WriteField("number", strconv.Itoa(data.Number)); err != nil {
		return nil, err
	}

	if err := form.WriteField("album", data.AlbumId); err != nil {
		return nil, err
	}

	if err := form.WriteField("artist", data.ArtistId); err != nil {
		return nil, err
	}

	if data.BestQualityFile.Content != nil {
		createFileField(form, "bestQualityFile", &data.BestQualityFile)
	}

	if data.MobileQualityFile.Content != nil {
		createFileField(form, "mobileQualityFile", &data.MobileQualityFile)
	}

	if data.CoverArt.Content != nil {
		createFileField(form, "coverArt", &data.CoverArt)
	}

	if err := form.Close(); err != nil {
		return nil, err
	}

	req, err := server.newReq("POST", "/tracks", &b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", form.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		return nil, errors.New("Request error: " + string(body))
	}

	var response types.ApiResponse[types.ApiPostTrackData]
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response.Data, nil
}
