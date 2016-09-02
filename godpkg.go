package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"bytes"
	"net/http"
)

func main() {
	fmt.Println("ready")

    if len(os.Args) < 3 {
        printUsage();
        return;
    }

    fileName := os.Args[1];
    apiUrl := os.Args[2];

	_, err := os.Stat(fileName);

    if err != nil {
        fmt.Println(fileName);
        fmt.Println(`Cant open file`);

        return;
    }

    proceed(fileName, apiUrl)
}


func printUsage() {
    fmt.Println(`
Usage:
    godpkg PATH_TO_DPKGLOG PATH_TO_ELASTICSEARCH");

Example:
    godpkg /var/log/dpkg.log http://localhost:9200/events-dpkg/packages
    `);
}


/**
 * Watching for file changes
 * Reporting changes to ElasticSearch
 */
func proceed(fileName string, apiUrl string) {
    fileInfo, _ := os.Stat(fileName);
	oldTime := fileInfo.ModTime()
	oldSize := fileInfo.Size()

	for {
		fileInfo, _ := os.Stat(fileName)

		newTime := fileInfo.ModTime();
        if oldTime == newTime {
    		time.Sleep(time.Second * 5)
            continue
        }

        if oldSize > fileInfo.Size() {
            //read from 0
            oldSize = 0
        }

        filePointer, _ := os.Open(fileName)
        filePointer.Seek(oldSize, 0)
        fileReader := bufio.NewReader(filePointer)
        for {
            line, error := fileReader.ReadString(10)

            if error == io.EOF {
                break
            }

            processLine(line, apiUrl)
        }

        oldTime = newTime
        oldSize = fileInfo.Size()
        filePointer.Close()
		time.Sleep(time.Second * 5)
	}

}

/**
 * matches only status string
func processLine(line string, apiUrl string) {
    fmt.Println(line)
	matchDate, _ := regexp.Compile(`^([0-9\-\:\ ]{19})`)
	matchStatus, _ := regexp.Compile(`status ([a-z]+) `)
	matchProject, _ := regexp.Compile(`status [a-z]+ ([a-z\d\:\.\-]+) `)
	matchBuild, _ := regexp.Compile(`2\:([0-9]+)\.`)
	matchVersion, _ := regexp.Compile(`[a-z\.] 2\:(.*)\+`)

    data := line //string(line)

    date := matchDate.FindStringSubmatch(data)
    status := matchStatus.FindStringSubmatch(data)
    project := matchProject.FindStringSubmatch(data)
    build := matchBuild.FindStringSubmatch(data)
    version := matchVersion.FindStringSubmatch(data)

    if len(date) != 2 || len(status) != 2 || len(project) != 2 || len(build) != 2 || len(version) != 2 {
        fmt.Println("Can't parse, skipping")
        return
    }

    dateUtc, _ := time.Parse(`2006-01-02 15:04:05 -0700`, fmt.Sprintf(`%s +0700`, date[1]))

    dateTimeUtc := dateUtc.UTC().Format("2006-01-02 15:04:05")

    bodyData := fmt.Sprintf(`{"@timestamp": "%s", "tags": ["%s", "%s"], "version": "%s", "build": %s, "project": "%s"}`, dateTimeUtc, status[1], project[1], version[1], build[1], project[1])
    client := &http.Client{}
    request, _ := http.NewRequest(`POST`, apiUrl, bytes.NewBufferString(bodyData))
    request.Header.Add(`Content-Type`, `application/json`)

    response, _ := client.Do(request)

    fmt.Println(response)
}
