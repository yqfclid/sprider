package main

import(
    "fmt"
    "net/http"
    "os"
    "github.com/axgle/mahonia"
    "github.com/PuerkitoBio/goquery"
    "regexp"
    "sync"
    "io"
)

type note struct{
    name string
    link string
}

type chapter struct{
    notename string
    title string
    link string
} 

func main() {
    var wg sync.WaitGroup
    url := "https://www.230book.com/dushixiaoshuo"
    pageForeach(url, &wg)
    wg.Wait()
    fmt.Println("all notes download success!")
}

func pageForeach(url string, wg *sync.WaitGroup){
    result := httpGet(url)
    defer result.Body.Close() 
    doc := getDoc(result.Body)
    notes := []*note{}
    noterawlist := doc.Find(".l").Find("ul")
    noterawlist.Find("li").Each(func(i int, s *goquery.Selection){
        noteA := s.Find(".s2").Find("a")
        link, exist := noteA.Attr("href")
        if exist == true{
            notename := convert2utf8(noteA.Text(), "gbk", "utf8")
            notes = append(notes, &note{name:notename, link:link})
        }
    }) 
    for _, i := range notes{
        wg.Add(1)
        go noteSprider(i, wg)
    }
    nextPage := doc.Find(".page_b").Find(".next")
    nextLink,exist := nextPage.Attr("href")
    if exist == true{
        pageForeach(nextLink, wg)
    }
}

func noteSprider(n *note, wg *sync.WaitGroup){
    defer wg.Done()
    fmt.Printf("start load note %v\n", n.name)
    f, err := os.OpenFile(n.name + ".txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
    if err != nil{
        fmt.Printf("open file error %v\n", err)
        return
    }
    result := httpGet(n.link)
    doc := getDoc(result.Body)
    chapters := []*chapter{}
    doc.Find("._chapter").Each(func(i int, s *goquery.Selection){
        s.Find("li").Each(func(i1 int, s1 *goquery.Selection){
            chapterA := s1.Find("a")
            path, exist := chapterA.Attr("href")
            if exist == true{
                title := convert2utf8(chapterA.Text(), "gbk", "utf8")
                link := n.link + path
                chapters = append(chapters, &chapter{title:title, link:link, notename:n.name})
            }
        })
    })
    offset, _  := f.Seek(0, os.SEEK_END)
    for _, chapterI := range chapters{
        title := "\n\n" + chapterI.title + "\n"
        content := getChapterContent(chapterI)
        writed, _ := f.WriteAt([]byte(title + content), offset)
        offset += int64(writed)
    }
    fmt.Printf("load note %v end\n", n.name)
    defer result.Body.Close()
    defer f.Close()
}


func getChapterContent(ch *chapter) string{
    fmt.Printf("start load chapter %v %v\n", ch.notename, ch.title)
    result := httpGet(ch.link)
    doc := getDoc(result.Body)
    content := ""
    var contentChapter *goquery.Selection
    contentBefore := doc.Find("#content")
    if contentBefore.Text() == ""{
        contentChapter = doc.Find("#chaptercontent")
    }else{
        contentChapter = contentBefore
    }
    chapterContent := convert2utf8(contentChapter.Text(), "gbk", "utf8")
    re, err := regexp.Compile("聽聽聽聽")
    if err != nil{
        fmt.Printf("re error %v", err)
        return ""
    }
    chapterContent = re.ReplaceAllString(chapterContent, "\n    ");
    content += chapterContent
    return content
}

func httpGet(url string)*http.Response{
    resp, err := http.Get(url)
    if  err != nil || resp.StatusCode != 200{
        fmt.Printf("http failed %v\n", err)
        return httpGet(url)
    }
    return resp
}

func getDoc(r io.Reader)*goquery.Document{
    doc, err := goquery.NewDocumentFromReader(r)
    if err != nil{
        fmt.Printf("getdoc failed %v\n", err)
        return getDoc(r)
    }
    return doc
}

func convert2utf8(src string, srcCode string, tagCode string) string {
    srcCoder := mahonia.NewDecoder(srcCode)
    srcResult := srcCoder.ConvertString(src)
    tagCoder := mahonia.NewDecoder(tagCode)
    _, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
    result := string(cdata)
    return result
}