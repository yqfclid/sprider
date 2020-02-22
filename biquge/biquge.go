package main

import(
    "fmt"
    "net/http"
    "os"
    "github.com/axgle/mahonia"
    "github.com/PuerkitoBio/goquery"
    "regexp"
    // "sync"
    "io"
    "strings"
    "net/url"
    "strconv"
)

type searchBook struct{
    key string
    name string
    lastUpdateTime string
    author string
    latestChap string
    sortNumber string
}

type note struct{
    name string
    link string
    prefix string
}

type chapter struct{
    notename string
    title string
    link string
} 

type spriderRes struct{
    key string
    result bool
}

const biqugeUrl string = "https://www.biqugela.com"

func main() {
    // keys := []string{}
    // keys = append(keys, "book_15734")
    // notesSprider(keys)


    i := 0
    for{
        keys := []string{}
        for j := 0;j < 50;j++{
            appendKey := "book_" + strconv.FormatInt(int64(i * 50 + j + 1), 10)
            keys = append(keys, appendKey)
        }
        result := notesSprider(keys)
        isFalse := false
        for k, v := range(result){
            if v == false{
                fmt.Printf("spride key %v failed\n", k)
                isFalse = true
            }
        }
        if isFalse{
            break
        }
        i++
    }


    // searchKey := "唐家三少"
    // searchBooks := noteSearch(searchKey)
    // for _,searchBook := range(searchBooks){
    //     fmt.Printf("sortNumber %v\n", searchBook.sortNumber)
    //     fmt.Printf("name%v\n", searchBook.name)
    //     fmt.Printf("author %v\n", searchBook.author)
    //     fmt.Printf("latestChap %v\n", searchBook.latestChap)
    //     fmt.Printf("link_key %v\n", searchBook.key)
    //     fmt.Printf("lastUpdateTime %v\n", searchBook.lastUpdateTime)
    // }
}

func noteSearch(searchKey string) []*searchBook{
    searchKeyEncoded := url.QueryEscape(searchKey)
    Url := biqugeUrl + "/search?keyword=" + searchKeyEncoded
    result := httpGet(Url)
    books := []*searchBook{}
    doc := getDoc(result.Body)
    doc.Find(".novelslist2").Find("ul").Find("li").Each(func(i int, s *goquery.Selection){
        if i > 0{
            book := &searchBook{}
            book.sortNumber = s.Find(".s1").Text()
            book.name = s.Find(".s2").Text()
            rawSearchKey, exist := s.Find(".s2").Find("a").Attr("href")
            if exist == true{
                book.key = strings.Split(rawSearchKey, "/")[1]
            }
            book.author = s.Find(".s4").Text()
            book.latestChap = s.Find(".s3").Text()
            book.lastUpdateTime = s.Find(".s6").Text()
            books = append(books, book)
        }
    })
    defer result.Body.Close()
    return books
}

func notesSprider(keys []string) map[string]bool{
    // var wg sync.WaitGroup
    resultMap := map[string]bool{}
    ch := make(chan spriderRes)
    for _, key := range keys{
        // wg.Add(1)
        // go noteSprider(key, &wg)
        go noteSprider(key, ch)
    }
    // wg.Wait()
    for i := 0;i < len(keys);i++{
        res := <- ch
        resultMap[res.key] = res.result
    }
    fmt.Println("all notes download end!")
    return resultMap
}

// func noteSprider(key string, wg *sync.WaitGroup){
    // defer wg.Done()
func noteSprider(key string, ch chan spriderRes){
    n := &note{link:biqugeUrl + "/" + key, prefix:biqugeUrl}
    result := httpGet(n.link)
    doc := getDoc(result.Body)
    chapters := []*chapter{}
    menuStr := n.name + "正文"
    realMenu := false
    n.name = doc.Find("#info").Find("h1").Text()
    fmt.Printf("start load note %v\n", n.name)
    f, err := os.OpenFile(n.name + ".txt", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
    if err != nil{
        fmt.Printf("open file error %v\n", err)
        return
    }
    doc.Find(".box_con").Find(".clearfix").Find("dd,dt").Each(func(i int, s *goquery.Selection){
        if realMenu{
            // fmt.Printf("load note %v end\n", s.Text())
            chapterA := s.Find("a")
            path, exist := chapterA.Attr("href")
            if exist == true{
                title := string(chapterA.Text())
                link := n.prefix + path
                chapters = append(chapters, &chapter{title:title, link:link, notename:n.name})
            }
        }else{
            re, _ := regexp.Match(menuStr, []byte(s.Text()))
            if re == true{
                realMenu = true
            }
        }
    })
    offset, _  := f.Seek(0, os.SEEK_END)
    for _, chapterI := range chapters{
        title := "\n\n" + chapterI.title + "\n"
        content := getChapterContent(chapterI)
        writed, _ := f.WriteAt([]byte(title + content), offset)
        offset += int64(writed)
    }
    defer result.Body.Close()
    defer f.Close()
    sprRes := spriderRes{key:key}
    if len(chapters) == 0{
        fmt.Printf("load note %v end\n", key)
        sprRes.result = false
        ch <- sprRes
        return 
    }
    fmt.Printf("load note %v with %v end\n", n.name, key)
    sprRes.result = true
    ch <- sprRes
}


func getChapterContent(ch *chapter) string{
    fmt.Printf("start load chapter %v %v %v\n", ch.notename, ch.title, ch.link)
    result := httpGet(ch.link)
    doc := getDoc(result.Body)
    content := ""
    doc.Find("#content").Find(".content_detail").Each(func(i int, s *goquery.Selection){
        content += "\n"
        content += string(s.Text())
    })
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