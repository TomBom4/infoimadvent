package routes

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"

	"github.com/hoffx/infoimadvent/config"
	"github.com/hoffx/infoimadvent/storage"
	macaron "gopkg.in/macaron.v1"
)

// Upload handles the route "/upload"
func Upload(ctx *macaron.Context, log *log.Logger, dStorer *storage.DocumentStorer) {
	defer ctx.HTML(200, "upload")

	ctx.Data["MinYL"] = config.Config.Grades.Min
	ctx.Data["MaxYL"] = config.Config.Grades.Max

	if ctx.Req.Method == "GET" {
		return
	}

	// parse trivial form values
	fValMinGrade := ctx.Req.FormValue("mingrade")
	if fValMinGrade == "" {
		fValMinGrade = "0"
	}
	fMinGrade, err := strconv.Atoi(fValMinGrade)
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
		return
	}
	fValMaxGrade := ctx.Req.FormValue("maxgrade")
	if fValMaxGrade == "" {
		fValMaxGrade = "0"
	}
	fMaxGrade, err := strconv.Atoi(fValMaxGrade)
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
		return
	}
	fValDay := ctx.Req.FormValue("day")
	if fValDay == "" {
		fValDay = "0"
	}
	fDay, err := strconv.Atoi(fValDay)
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
		return
	}
	fSolution := ctx.Req.FormValue("solution")
	fType := ctx.Req.FormValue("type")

	// save trivial form values
	ctx.Data["Day"] = fDay
	ctx.Data["MinGrade"] = fMinGrade
	ctx.Data["MaxGrade"] = fMaxGrade
	ctx.Data[fSolution] = true

	var docType int
	switch fType {
	case "About":
		docType = storage.About
		ctx.Data["About"] = true
		fMinGrade = 0
		fMaxGrade = 0
		fSolution = ""
		fDay = 0
	case "Terms of Service":
		docType = storage.ToS
		ctx.Data["ToS"] = true
		fMinGrade = 0
		fMaxGrade = 0
		fSolution = ""
		fDay = 0
	default:
		docType = storage.Quest
		ctx.Data["Quest"] = true
	}

	solution, err := solutionToInt(fSolution)
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
		return
	}

	// load and save form files
	fMd, _, err := ctx.Req.FormFile("md")
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
		log.Println(err)
		return
	}
	defer fMd.Close()

	f, err := ioutil.TempFile(config.Config.FileSystem.MDStoragePath, "document")
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrFS)
		log.Println(err)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, fMd)
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrFS)
		log.Println(err)
		return
	}

	fAssets, _, err := ctx.Req.FormFile("assets")
	if err != nil {
		ctx.Data["Error"] = ctx.Tr(ErrNoAssets)
		if err.Error() != "http: no such file" {
			log.Println(err)
		}
	} else {
		defer fAssets.Close()
		buf := new(bytes.Buffer)
		length, err := buf.ReadFrom(fAssets)
		if err != nil {
			ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
			log.Println(err)
		}

		reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), length)
		if err != nil {
			ctx.Data["Error"] = ctx.Tr(ErrIllegalInput)
			log.Println(err)
			return
		}
		err = unzipAndSave(*reader, config.Config.FileSystem.AssetsStoragePath+"/"+path.Base(f.Name()))
		if err != nil {
			ctx.Data["Error"] = ctx.Tr(ErrFS)
			log.Println(err)
			return
		}
	}

	// create db entries

	for i := fMinGrade; i <= fMaxGrade; i++ {
		doc := storage.Document{
			Path:     f.Name(),
			Grade:    i,
			Day:      fDay,
			Solution: solution,
			Type:     docType,
		}
		oldDoc, err := dStorer.Get(map[string]interface{}{"grade": i, "day": fDay, "type": docType})
		if err != nil {
			ctx.Data["Error"] = ctx.Tr(ErrDB)
			log.Println(err)
			return
		}
		if oldDoc.Path == "" {
			err = dStorer.Create(doc)
			if err != nil {
				ctx.Data["Error"] = ctx.Tr(ErrDB)
				log.Println(err)
				return
			}
		} else {
			err = dStorer.Put(doc)
			if err != nil {
				ctx.Data["Error"] = ctx.Tr(ErrDB)
				log.Println(err)
				return
			}
		}
	}
}

func solutionToInt(sol string) (solution int, err error) {
	switch sol {
	case "":
		solution = storage.None
	case "A":
		solution = storage.A
	case "B":
		solution = storage.B
	case "C":
		solution = storage.C
	case "D":
		solution = storage.D
	default:
		err = errors.New(ErrIllegalInput)
	}
	return
}

func unzipAndSave(reader zip.Reader, target string) error {

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}
