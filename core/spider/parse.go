package spider

import (
	"errors"
	"fmt"
	"github.com/hunterhug/marmot/expert"
	"github.com/hunterhug/marmot/util/goquery"
	"regexp"
	"strings"
	"taotie/core/util"
)

func Is404(content []byte) bool {
	doc, _ := expert.QueryBytes(content)
	text := doc.Find("title").Text()
	if strings.Contains(text, "Page Not Found") {
		return true
	}
	if strings.Contains(text, "404") {
		return true
	}
	//uk
	if strings.Contains(string(content), "The Web address you entered is not a functioning page on our site") {
		return true
	}
	//de
	if strings.Contains(string(content), "Suchen Sie bestimmte Informationen") {
		return true
	}
	if strings.Contains(string(content), "Suchen Sie etwas bestimmtes") {
		return true
	}
	return false
}

func IsRobot(content []byte) (s string) {
	doc, _ := expert.QueryBytes(content)
	text := doc.Find("title").Text()
	//if text == "" {
	//	return "empty"
	//}

	if strings.Contains(text, "Sorry! Something went wrong!") {
		return "sorry"
	}

	// uk usa
	if strings.Contains(text, "Robot Check") {
		return "robot"
	}
	//jp
	if strings.Contains(text, "CAPTCHA") {
		return "robot"
	}
	//de
	if strings.Contains(text, "Bot Check") {
		return "robot"
	}
	return
}

func TooSortSizes(data []byte, sizes float64) error {
	if float64(len(data))/1000 < sizes {
		return errors.New(fmt.Sprintf("FileSize:%d bytes,%d kb < %f kb dead too sort", len(data), len(data)/1000, sizes))
	}
	return nil
}

func ParseList(content []byte, other bool) ([]map[string]string, error) {
	if other {
		return ParseListOtherType(content)
	}

	returnMap := make([]map[string]string, 0)

	doc, _ := expert.QueryBytes(content)
	goodClass := ".zg-item-immersion"
	doc.Find(goodClass).Each(func(i int, node *goquery.Selection) {
		dudu := node.Find("a")
		text, exist := dudu.Attr("href")
		if exist {
			temp := map[string]string{}
			if strings.Contains(text, "/dp/") {
				t1 := strings.Split(text, "/dp/")
				if len(t1) != 2 {
					return
				}
				t2 := strings.Split(t1[1], "/")
				temp["asin"] = t2[0]
				if temp["asin"] == "" {
					return
				}

			} else if strings.Contains(text, "/product/") {
				temp := map[string]string{}
				t1 := strings.Split(text, "/product/")
				if len(t1) != 2 {
					return
				}
				t2 := strings.Split(t1[1], "/")
				temp["asin"] = t2[0]
				if temp["asin"] == "" {
					return
				}

			} else {
				return
			}
			imag := dudu.Find("img")
			temp["title"], _ = imag.Attr("alt")
			temp["img"], _ = imag.Attr("src")

			score := strings.TrimSpace(node.Find(".a-icon-row").Text())
			temp["reviews"] = "0"
			temp["score"] = "0"

			scoreTemp := strings.Split(score, "star")
			switch len(scoreTemp) {
			case 1:
				temp["score"] = strings.Replace(scoreTemp[0], "out of", "", -1)
			case 2:
				temp["score"] = strings.Replace(scoreTemp[0], "out of", "", -1)
				temp["reviews"] = strings.TrimSpace(strings.Replace(scoreTemp[1], "s", "", -1))
			}
			temp["score"] = strings.TrimSpace(strings.Replace(temp["score"], "5", "", -1))
			temp["reviews"] = strings.Replace(temp["reviews"], ",", "", -1)

			temp["small_rank"] = strings.TrimSpace(strings.Replace(strings.Replace(node.Find(".zg-badge-text").Text(), ".", "", -1), "#", "", -1))

			if temp["reviews"] == "" {
				temp["reviews"] = "0"
			}
			if temp["score"] == "" {
				temp["score"] = "0"
			}

			temp["price"] = strings.Replace(strings.TrimSpace(node.Find(".a-color-price").Text()), "$", "", -1)
			isPrime := node.Find(".a-icon-prime").Size() > 0
			temp["is_prime"] = fmt.Sprintf("%v", isPrime)
			returnMap = append(returnMap, temp)
			return
		}
	})

	if len(returnMap) == 0 {
		return nil, errors.New("parse get null")
	}
	return returnMap, nil
}

func ParseListOtherType(content []byte) ([]map[string]string, error) {
	returnMap := make([]map[string]string, 0)

	doc, _ := expert.QueryBytes(content)
	goodClass := ".s-result-item"
	doc.Find(goodClass).Each(func(i int, node *goquery.Selection) {
		asin, exist := node.Attr("data-asin")
		if !exist {
			return
		}

		temp := map[string]string{}
		temp["small_rank"], exist = node.Attr("data-index")
		if exist {
			i, err := util.SInt64(temp["small_rank"])
			if err == nil {
				temp["small_rank"] = fmt.Sprintf("%d", i+1)
			}
		}
		temp["asin"] = asin
		imag := node.Find("img")
		temp["title"], _ = imag.Attr("alt")
		temp["img"], _ = imag.Attr("src")

		score := node.Find(".a-section .a-row span .a-declarative").Text()
		score = strings.TrimSpace(score)

		scoreArray := strings.Split(score, " out of")
		if len(scoreArray) >= 2 {
			temp["score"] = strings.TrimSpace(scoreArray[0])
		}

		reviews := node.Find(".a-section .a-row span .a-link-normal").Text()
		reviews = strings.TrimSpace(reviews)
		temp["reviews"] = strings.TrimSpace(reviews)
		price := node.Find(".a-section .a-row .a-price").Text()
		price = strings.TrimSpace(price)
		priceArray := strings.Split(price, "$")
		if len(priceArray) >= 2 {
			temp["price"] = priceArray[1]
		}

		returnMap = append(returnMap, temp)
	})

	if len(returnMap) == 0 {
		return nil, errors.New("parse get null")
	}
	return returnMap, nil
}

type AsinDetail struct {
	Asin       string
	Title      string
	BigName    string
	IsStock    bool
	IsFba      bool
	IsAwsSold  bool
	SoldBy     string
	SoldById   string
	Img        string
	IsPrime    bool
	Price      float64
	Reviews    int64
	Score      float64
	Describe   string
	BigRank    int64
	RankDetail string
}

func ParseDetail(content []byte) (detail *AsinDetail, err error) {
	detail = new(AsinDetail)
	doc, _ := expert.QueryBytes(content)

	titleTrip := "Amazon.com:"

	//detailBulletsWrapper_feature_div
	title := strings.Replace(doc.Find("title").Text(), titleTrip, "", -1)

	if title == "" {
		return nil, errors.New("title empty")
	}
	bigName := "null"
	temp := strings.Split(title, ":")
	tempL := len(temp)
	if tempL >= 2 {
		bigName = strings.TrimSpace(temp[tempL-1])
		title = strings.Join(temp[0:tempL-1], ":")
	} else {
		temp := strings.Split(title, " at ")
		tempL := len(temp)
		if tempL >= 2 {
			bigName = strings.TrimSpace(temp[tempL-1])
			title = strings.Join(temp[0:tempL-1], " at ")
		}
	}

	detail.Title = strings.TrimSpace(title)
	detail.BigName = bigName

	detail.Img, _ = doc.Find("#imgTagWrapperId img").Attr("data-old-hires")
	if detail.Img == "" {
		imgStr, _ := doc.Find("#imgTagWrapperId img").Attr("data-a-dynamic-image")
		imgArray := strings.Split(imgStr, "\":")
		if len(imgArray) >= 2 {
			imgTemp := strings.Replace(imgArray[0], "{\"", "", -1)
			if strings.HasPrefix(imgTemp, "http") {
				detail.Img = imgTemp
			}
		}
	}

	inStock := doc.Find("#availability span").Text()
	if strings.Contains(inStock, "Currently unavailable.") {

	} else {
		detail.IsStock = true
	}

	merchantInfo := strings.TrimSpace(doc.Find("#merchant-info").Text())
	if strings.Contains(merchantInfo, "Ships from and sold by Amazon.com.") {
		detail.IsFba = true
		detail.IsAwsSold = true
		detail.SoldBy = "Amazon.com"
	} else {
		if strings.Contains(merchantInfo, "Fulfilled by Amazon.") {
			detail.IsFba = true
		}

		detail.SoldById, _ = doc.Find("#merchant-info #seller-popover-information").Attr("data-merchant-id")
		detail.SoldBy = doc.Find("#merchant-info #sellerProfileTriggerId").Text()
	}

	detail.Describe = fmt.Sprintf("<p>%s</p>", strings.TrimSpace(doc.Find("#productDescription p").Text()))

	review := strings.TrimSpace(doc.Find("#prodDetails #acrCustomerReviewText").Text())
	detail.Reviews, _ = util.SInt64(strings.Replace(review, " ratings", "", -1))

	score := strings.TrimSpace(doc.Find("#prodDetails .a-icon-star").Text())
	detail.Score, _ = util.SFloat64(strings.Replace(score, " out of 5 stars", "", -1))
	// descriptionAndDetails
	//prodDetails

	rankStr := strings.TrimSpace(doc.Find("#prodDetails").Text())
	r, _ := regexp.Compile(`#([,\d]{1,10})[\s]{0,1}[A-Za-z0-9]{0,6} in ([^#;)(\n]{2,30})[\s\n]{0,1}[(]{0,1}`)
	god := r.FindAllStringSubmatch(rankStr, -1)
	if len(god) > 0 {
		if len(god[0]) >= 2 {
			detail.BigRank, _ = util.SInt64(strings.Replace(god[0][1], ",", "", -1))
		}
	}

	i := 0
	for _, v := range god {
		if i == 0 {
			detail.RankDetail = strings.Replace(strings.Replace(v[0], " (", "", -1), "\n", "", -1)
			i = i + 1
			continue
		}
		detail.RankDetail = detail.RankDetail + "\n" + strings.Replace(strings.Replace(v[0], " (", "", -1), "\n", "", -1)
	}

	price := doc.Find(".priceBlockBuyingPriceString").Text()
	if price != "" {
		priceArray := strings.Split(price, " - ")
		if len(priceArray) > 0 {
			detail.Price, _ = util.SFloat64(strings.Replace(priceArray[0], "$", "", -1))
		}
	}

	table := make([]string, 0)
	doc.Find("#prodDetails table tbody tr").Each(func(i int, selection *goquery.Selection) {
		th := strings.TrimSpace(selection.Find("th").Text())
		td := strings.TrimSpace(selection.Find("td").Text())

		if strings.Contains(th, "Customer Reviews") {
			return
		}
		table = append(table, fmt.Sprintf(`<tr><th>%s</th><td>%s</td></tr>`, th, td))
	})
	if len(table) > 0 {
		if detail.Describe == "" {
			detail.Describe = fmt.Sprintf("<table>%s</table>", strings.Join(table, ""))
		} else {
			detail.Describe = fmt.Sprintf("%s<br/><table>%s</table>", detail.Describe, strings.Join(table, ""))
		}
	}
	return
}
