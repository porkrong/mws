package submit

import "encoding/xml"

type xmlHeader struct {
	Header struct {
		DocumentVersion    string
		MerchantIdentifier string
	}
	MessageType string
}

func NewProductXml() (obj *ProductXml) {
	obj = new(ProductXml)
	obj.Header.DocumentVersion = "1.01"
	obj.MessageType = "Product"
	return
}

func NewProductImgXml() (obj *ProductImgXml) {
	obj = new(ProductImgXml)
	obj.Header.DocumentVersion = "1.01"
	obj.MessageType = "ProductImage"
	return
}

func NewProductPriceXml() (obj *ProductPriceXml) {
	obj = new(ProductPriceXml)
	obj.Header.DocumentVersion = "1.01"
	obj.MessageType = "Price"
	return
}

func NewProductQuantityXml() (obj *ProductPriceXml) {
	obj = new(ProductPriceXml)
	obj.Header.DocumentVersion = "1.01"
	obj.MessageType = "Inventory"
	return
}

type ProductXml struct {
	XMLName xml.Name `xml:"AmazonEnvelope"`
	xmlHeader
	Message struct {
		MessageID     string
		OperationType string
		Product       struct {
			SKU             string
			ProductTaxCode  string
			LaunchDate      string
			DescriptionData struct {
				Title                 string
				Brand                 string
				Description           string
				RecommendedBrowseNode string
			}
		}
	}
}

type ProductImgXml struct {
	xmlHeader
}

type ProductPriceXml struct {
	xmlHeader
}

type ProductQuantityXml struct {
	xmlHeader
}
