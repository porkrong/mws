package main

import (
	"io/ioutil"
	"net/http"
	"net/url"

	"time"

	"encoding/json"
	"fmt"
)

func main() {

	for j := 1; j < 4; j++ {
		time.Sleep(time.Second * 2)
		list := []map[string]string{}
		for i := 0; i < 100; i++ {
			data := make(map[string]string)
			data["feed_product_type"] = "sportinggoods"
			data["item_sku"] = fmt.Sprintf("%v", 10162513343443+i)
			data["manufacturer"] = "manufacturer"
			data["brand_name"] = "Nescafe"
			data["item_name"] = "test-title"
			data["external_product_id"] = "079346140764"
			data["external_product_id_type"] = "UPC"
			data["recommended_browse_nodes"] = "1981010031"
			data["standard_price"] = "12.32"
			data["quantity"] = "23"
			data["main_image_url"] = "http://www.companyname.com/images/1250.main.jpg"
			data["product_description"] = "product_description"
		}

		b, _ := json.Marshal(list)
		param := make(map[string]string)
		param["auth_id"] = "1"
		param["sign"] = "1"
		param["create_time"] = "1343423"
		param["data"] = string(b)
		u := "http://127.0.0.1:8080"
		b, err := curl_post(u, param, time.Second*3)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(b))
	}

}

func curl_post(u string, param map[string]string, duration time.Duration) (b []byte, err error) {
	client := &http.Client{
		Timeout: duration,
	}
	p := url.Values{}
	for k, v := range param {
		p[k] = []string{v}
	}
	resp, err := client.PostForm(u, p)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	b, err = ioutil.ReadAll(resp.Body)
	return
}
