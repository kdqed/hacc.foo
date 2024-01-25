package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"
)

const port string = ":2511"
var tmplt *template.Template

type itemsList struct {
    Items []map[string]any
    NextPageStartId int
    Error bool
}

func findIndex(arr []int, value int, fallback int) int {
    for i := 0; i < len(arr); i++ {
        if arr[i]==value {
            return i
        }
    }
    return fallback
}

func getItemsData(itemIds []int) []map[string]any {
    var data []map[string]any
    for i := 0; i < len(itemIds); i++ {
        url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/item/%d.json", itemIds[i])
        
        var target map[string]any
        
        resp, err := http.Get(url)
        if err!=nil {
            log.Println(url)
            log.Println(err)
            continue
        }
        
        defer resp.Body.Close()
        body, err := ioutil.ReadAll(resp.Body)
        if err!=nil {
            log.Println(url)
            log.Println(err)
            continue
        }
        
        d := json.NewDecoder(bytes.NewReader(body))
        d.UseNumber()
        d.Decode(&target)
        
        timestampInt, err := target["time"].(json.Number).Int64()
        if err!=nil {
            log.Println(err)
            continue
        }
        timestampObj := time.Unix(timestampInt, 0) 
        timeStr := strings.Split(time.Since(timestampObj).String(), ".")[0] + "s"
        displayTime := ""
        for _, c := range timeStr {
            displayTime += string(c)
            if c < '0' || c > '9' {
                break
            }
        }
        target["displayTime"] = displayTime
        
        if target["url"]==nil {
            idInt, _ := target["id"].(json.Number).Int64()
            target["url"] = fmt.Sprintf("https://news.ycombinator.com/item?id=%d", idInt) 
        }
        data = append(data, target)
        
    }
    return data
}

func getItemsList(storyType string, startId int, limit int) itemsList {
    var result itemsList
    result.Error = true
    result.Items = []map[string]any{}
    result.NextPageStartId = -1
    
    url := fmt.Sprintf("https://hacker-news.firebaseio.com/v0/%sstories.json", storyType)
    resp, err := http.Get(url)
    if err!=nil {
        return result
    }
    
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err!=nil {
        return result
    }
    
    var storyIds []int
    _ = json.Unmarshal(body, &storyIds)
    
    startIndex := 0
    if startId!=-1 {
        startIndex = findIndex(storyIds, startId, len(storyIds))
    }
    
    endIndex := startIndex + limit
    if startIndex >= len(storyIds) {
        startIndex = 0
        endIndex = 0
    } else if endIndex > len(storyIds) {
        endIndex = len(storyIds)
    }
    
    nextPageStartId := -1
    if endIndex > 0 &&  endIndex < len(storyIds) {
        nextPageStartId = storyIds[endIndex]
    }
    
    result.Error = false
    result.Items = getItemsData(storyIds[startIndex:endIndex])
    result.NextPageStartId = nextPageStartId
    
    return result
}

func handler(w http.ResponseWriter, r *http.Request) {
    tmplt, _ = template.ParseGlob("templates/*")
    
    storyType := strings.Trim(r.URL.Path, "/")
    if storyType=="" {
        storyType = "top"
    }
    
    startIdParam := r.FormValue("startId")
    startId := -1
    if startIdParam!="" {
        v, _ := strconv.Atoi(startIdParam)
        startId = v
    }
    
    result := getItemsList(storyType, startId, 5)
    context := [2]any{storyType, result} 
    
    hxRequest := r.Header.Get("HX-Request")
    var err any
    if hxRequest=="true" {
        err = tmplt.ExecuteTemplate(w, "list.html", context)
    } else {
        err = tmplt.ExecuteTemplate(w, "_wrapper.html", context)
    }
    
    if err != nil {
        log.Println(err)
    }
}

func main() {
    log.Println("Starting Server on Port", port)
    
    fs := http.FileServer(http.Dir("./static"))
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    
    http.HandleFunc("/", handler)
    http.HandleFunc("/top", handler)
    http.HandleFunc("/new", handler)
    http.HandleFunc("/best", handler)
    http.HandleFunc("/ask", handler)
    http.HandleFunc("/show", handler)
    http.HandleFunc("/job", handler)
    
    err := http.ListenAndServe(port, nil)
    if err != nil {
        log.Fatal(err)
    }
}
