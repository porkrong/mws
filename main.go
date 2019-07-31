package main

import (
	"encoding/json"
	"fmt"
	"github.com/svvu/gomws/mws"
	"mws/submit"
	"time"
)

func main() {

	//fmt.Println(string(body))
	//h := md5.New()
	//h.Write([]byte(body))
	//str := base64.StdEncoding.EncodeToString(h.Sum(nil))
	//
	//fmt.Println(str)
	//
	//fmt.Println(len("VEkkqSnZ3m5DVkW7ygQW6Q=="))
	//return
	config := mws.Config{
		SellerId:  "",
		AuthToken: "",
		Region:    "",
		AccessKey: "",
		SecretKey: "",
	}

	cli, err := submit.NewClient(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	data := make(map[string]string)
	data["feed_product_type"] = "sportinggoods"
	data["item_sku"] = "10162513343443"
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
	//list[0] = "sportinggoods"
	//list[1] = "10162513343443"
	//list[2] = "Nescafe"
	//list[3] = "test-title"
	//list[4] = "079346140764"
	//list[5] = "UPC"
	//list[6] = "1981010031"
	//list[7] = "12.32"
	//list[8] = "23"
	//list[9] = "http://www.companyname.com/images/1250.main.jpg"
	//list[10] = "product_description"
	//

	FeedSubmissionId_queue := make(chan string, 1000)
	//开启一个协程去获取执行结果
	go func() {
		for {
			//从管道中取出需要查询的
			FeedSubmissionId := <-FeedSubmissionId_queue
			fmt.Println("开始执行GetFeedSubmissionResult，当前参数为", FeedSubmissionId)
			rsp, err := cli.GetFeedSubmissionResult(FeedSubmissionId)
			if err != nil {
				fmt.Println("GetFeedSubmissionResult err ", err)
				return
			}

			if rsp["status"].(bool) == false {
				fmt.Println("当前FeedSubmissionId为", FeedSubmissionId, ",执行错误结果为:", rsp["error_content"])
				//并且写入到数据库中记录当前错误的信息
				continue
			}
			//获取到报告id
			ProcessingReport := rsp["ProcessingReport"].(string)
			fmt.Println("报告id为", ProcessingReport)
		}
	}()

	rsp, err := cli.SubmitTpl([]map[string]string{data})
	if err != nil {
		fmt.Println("SubmitTpl err ", err)
		return
	}
	fmt.Println(rsp)
	if rsp["status"].(bool) == false {
		fmt.Println(rsp["error_content"])
		return
	}
	FeedSubmissionId := rsp["FeedSubmissionId"].(string)
	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			fmt.Println("开始执行GetFeedSubmissionList,参数为", FeedSubmissionId)
			rsp, err = cli.GetFeedSubmissionList(FeedSubmissionId)
			if err != nil {
				fmt.Println("GetFeedSubmissionList err ", err)
				continue
			}

			if rsp["status"].(bool) == false {
				fmt.Println(rsp["error_content"])
				continue
			}

			//判断是否完成,未完成则继续
			FeedProcessingStatus := rsp["FeedProcessingStatus"].(string)
			switch FeedProcessingStatus {
			case "_DONE_":
				//放到管道中执行，获取结果
				fmt.Println("GetFeedSubmissionList执行完成")
				FeedSubmissionId_queue <- rsp["FeedSubmissionId"].(string)
			case "_CANCELLED_":
				//请求因严重错误而中止。
				break
			case "_IN_SAFETY_NET_":
				//请求正在处理，但系统发现上传数据可能包含潜在错误（例如，请求将删除卖家账户中的所有库存）。
				// 亚马逊卖家支持团队将联系卖家，以确认是否应处理该上传数据。
			default:
				fmt.Println("当前GetFeedSubmissionList的获取的状态为", FeedProcessingStatus)
				continue
			}

			NextToken := rsp["NextToken"].(string)
			if FeedProcessingStatus == "_DONE_" && NextToken == "" {
				goto This
			}
			if NextToken != "" {
				t := time.NewTicker(time.Second)
				defer t.Stop()
				for {
					select {
					case <-t.C:
						rsp, err = cli.GetFeedSubmissionListByNextToken(NextToken)
						if err != nil {
							fmt.Println("GetFeedSubmissionListByNextToken err ", err)
							continue
						}
						if rsp["status"].(bool) == false {
							fmt.Println(rsp["error_content"])
							continue
						}

						//判断是否完成,未完成则继续
						FeedProcessingStatus := rsp["FeedProcessingStatus"].(string)
						switch FeedProcessingStatus {
						case "_DONE_":
							//放到管道中执行，获取结果
							FeedSubmissionId_queue <- rsp["FeedSubmissionId"].(string)
						case "_CANCELLED_":
							//请求因严重错误而中止。
							break
						case "_IN_SAFETY_NET_":
							//请求正在处理，但系统发现上传数据可能包含潜在错误（例如，请求将删除卖家账户中的所有库存）。
							// 亚马逊卖家支持团队将联系卖家，以确认是否应处理该上传数据。
						default:
							continue
						}
						//已经完成，继续判断是否有NextToken，有就重复上一步，没有则跳出当前循环
						NextToken = rsp["NextToken"].(string)
						if NextToken == "" {
							break
						}

					}
				}
			}

		}

	}
This:
	fmt.Println(rsp)

	select {}

}
