package main

import (
	"encoding/xml"
	"github.com/DDRBoxman/go-amazon-product-api"
	"log"
	"net/http"
	"os"
)

func getAmazonAffiliateLink(isbn string) (amazonLink string) {
	associateTag := os.Getenv("AMAZON_TAG")
	if associateTag == "" {
		return
	}
	var api amazonproduct.AmazonProductAPI
	api.AccessKey = os.Getenv("AMAZON_KEY")
	api.SecretKey = os.Getenv("AMAZON_SECRET")
	api.Host = "webservices.amazon.com"
	api.AssociateTag = associateTag
	api.Client = &http.Client{} // optional

	result, err := api.ItemSearchByKeyword(isbn, 1)
	if err != nil {
		log.Output(0, err.Error())
		return
	}

	var response Amzn_ItemSearchResponse
	if err := xml.Unmarshal([]byte(result), &response); err != nil {
		log.Output(0, err.Error())
		return
	}
	return response.Amzn_Items.Amzn_Item.Amzn_DetailPageURL.Text
}

type Amzn_root struct {
	Amzn_ItemSearchResponse Amzn_ItemSearchResponse `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemSearchResponse,omitempty" json:"ItemSearchResponse,omitempty"`
}

type Amzn_ASIN struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ASIN,omitempty" json:"ASIN,omitempty"`
}

type Amzn_Argument struct {
	Attr_Name  string   `xml:" Name,attr"  json:",omitempty"`
	Attr_Value string   `xml:" Value,attr"  json:",omitempty"`
	XMLName    xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Argument,omitempty" json:"Argument,omitempty"`
}

type Amzn_Arguments struct {
	Amzn_Argument []Amzn_Argument `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Argument,omitempty" json:"Argument,omitempty"`
	XMLName       xml.Name        `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Arguments,omitempty" json:"Arguments,omitempty"`
}

type Amzn_Author struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Author,omitempty" json:"Author,omitempty"`
}

type Amzn_Binding struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Binding,omitempty" json:"Binding,omitempty"`
}

type Amzn_Code struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Code,omitempty" json:"Code,omitempty"`
}

type Amzn_Content struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Content,omitempty" json:"Content,omitempty"`
}

type Amzn_Description struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Description,omitempty" json:"Description,omitempty"`
}

type Amzn_DetailPageURL struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 DetailPageURL,omitempty" json:"DetailPageURL,omitempty"`
}

type Amzn_EISBN struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 EISBN,omitempty" json:"EISBN,omitempty"`
}

type Amzn_EditorialReview struct {
	Amzn_Content          Amzn_Content          `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Content,omitempty" json:"Content,omitempty"`
	Amzn_IsLinkSuppressed Amzn_IsLinkSuppressed `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 IsLinkSuppressed,omitempty" json:"IsLinkSuppressed,omitempty"`
	Amzn_Source           Amzn_Source           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Source,omitempty" json:"Source,omitempty"`
	XMLName               xml.Name              `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 EditorialReview,omitempty" json:"EditorialReview,omitempty"`
}

type Amzn_EditorialReviews struct {
	Amzn_EditorialReview Amzn_EditorialReview `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 EditorialReview,omitempty" json:"EditorialReview,omitempty"`
	XMLName              xml.Name             `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 EditorialReviews,omitempty" json:"EditorialReviews,omitempty"`
}

type Amzn_Error struct {
	Amzn_Code    Amzn_Code    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Code,omitempty" json:"Code,omitempty"`
	Amzn_Message Amzn_Message `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Message,omitempty" json:"Message,omitempty"`
	XMLName      xml.Name     `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Error,omitempty" json:"Error,omitempty"`
}

type Amzn_Errors struct {
	Amzn_Error Amzn_Error `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Error,omitempty" json:"Error,omitempty"`
	XMLName    xml.Name   `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Errors,omitempty" json:"Errors,omitempty"`
}

type Amzn_Format struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Format,omitempty" json:"Format,omitempty"`
}

type Amzn_HTTPHeaders struct {
	Amzn_Header Amzn_Header `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Header,omitempty" json:"Header,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 HTTPHeaders,omitempty" json:"HTTPHeaders,omitempty"`
}

type Amzn_Header struct {
	Attr_Name  string   `xml:" Name,attr"  json:",omitempty"`
	Attr_Value string   `xml:" Value,attr"  json:",omitempty"`
	XMLName    xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Header,omitempty" json:"Header,omitempty"`
}

type Amzn_Height struct {
	Attr_Units string   `xml:" Units,attr"  json:",omitempty"`
	Text       string   `xml:",chardata" json:",omitempty"`
	XMLName    xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
}

type Amzn_HiResImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 HiResImage,omitempty" json:"HiResImage,omitempty"`
}

type Amzn_ImageSet struct {
	Attr_Category       string              `xml:" Category,attr"  json:",omitempty"`
	Amzn_HiResImage     Amzn_HiResImage     `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 HiResImage,omitempty" json:"HiResImage,omitempty"`
	Amzn_LargeImage     Amzn_LargeImage     `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 LargeImage,omitempty" json:"LargeImage,omitempty"`
	Amzn_MediumImage    Amzn_MediumImage    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 MediumImage,omitempty" json:"MediumImage,omitempty"`
	Amzn_SmallImage     Amzn_SmallImage     `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SmallImage,omitempty" json:"SmallImage,omitempty"`
	Amzn_SwatchImage    Amzn_SwatchImage    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SwatchImage,omitempty" json:"SwatchImage,omitempty"`
	Amzn_ThumbnailImage Amzn_ThumbnailImage `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ThumbnailImage,omitempty" json:"ThumbnailImage,omitempty"`
	Amzn_TinyImage      Amzn_TinyImage      `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 TinyImage,omitempty" json:"TinyImage,omitempty"`
	XMLName             xml.Name            `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ImageSet,omitempty" json:"ImageSet,omitempty"`
}

type Amzn_ImageSets struct {
	Amzn_ImageSet Amzn_ImageSet `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ImageSet,omitempty" json:"ImageSet,omitempty"`
	XMLName       xml.Name      `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ImageSets,omitempty" json:"ImageSets,omitempty"`
}

type Amzn_IsLinkSuppressed struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 IsLinkSuppressed,omitempty" json:"IsLinkSuppressed,omitempty"`
}

type Amzn_IsValid struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 IsValid,omitempty" json:"IsValid,omitempty"`
}

type Amzn_Item struct {
	Amzn_ASIN             Amzn_ASIN             `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ASIN,omitempty" json:"ASIN,omitempty"`
	Amzn_DetailPageURL    Amzn_DetailPageURL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 DetailPageURL,omitempty" json:"DetailPageURL,omitempty"`
	Amzn_EditorialReviews Amzn_EditorialReviews `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 EditorialReviews,omitempty" json:"EditorialReviews,omitempty"`
	Amzn_ImageSets        Amzn_ImageSets        `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ImageSets,omitempty" json:"ImageSets,omitempty"`
	Amzn_ItemAttributes   Amzn_ItemAttributes   `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemAttributes,omitempty" json:"ItemAttributes,omitempty"`
	Amzn_ItemLinks        Amzn_ItemLinks        `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemLinks,omitempty" json:"ItemLinks,omitempty"`
	Amzn_LargeImage       Amzn_LargeImage       `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 LargeImage,omitempty" json:"LargeImage,omitempty"`
	Amzn_MediumImage      Amzn_MediumImage      `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 MediumImage,omitempty" json:"MediumImage,omitempty"`
	Amzn_SmallImage       Amzn_SmallImage       `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SmallImage,omitempty" json:"SmallImage,omitempty"`
	XMLName               xml.Name              `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Item,omitempty" json:"Item,omitempty"`
}

type Amzn_ItemAttributes struct {
	Amzn_Author          Amzn_Author          `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Author,omitempty" json:"Author,omitempty"`
	Amzn_Binding         Amzn_Binding         `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Binding,omitempty" json:"Binding,omitempty"`
	Amzn_EISBN           Amzn_EISBN           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 EISBN,omitempty" json:"EISBN,omitempty"`
	Amzn_Format          Amzn_Format          `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Format,omitempty" json:"Format,omitempty"`
	Amzn_Label           Amzn_Label           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Label,omitempty" json:"Label,omitempty"`
	Amzn_Languages       Amzn_Languages       `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Languages,omitempty" json:"Languages,omitempty"`
	Amzn_Manufacturer    Amzn_Manufacturer    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Manufacturer,omitempty" json:"Manufacturer,omitempty"`
	Amzn_NumberOfPages   Amzn_NumberOfPages   `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 NumberOfPages,omitempty" json:"NumberOfPages,omitempty"`
	Amzn_ProductGroup    Amzn_ProductGroup    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ProductGroup,omitempty" json:"ProductGroup,omitempty"`
	Amzn_ProductTypeName Amzn_ProductTypeName `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ProductTypeName,omitempty" json:"ProductTypeName,omitempty"`
	Amzn_PublicationDate Amzn_PublicationDate `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 PublicationDate,omitempty" json:"PublicationDate,omitempty"`
	Amzn_Publisher       Amzn_Publisher       `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Publisher,omitempty" json:"Publisher,omitempty"`
	Amzn_ReleaseDate     Amzn_ReleaseDate     `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ReleaseDate,omitempty" json:"ReleaseDate,omitempty"`
	Amzn_Studio          Amzn_Studio          `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Studio,omitempty" json:"Studio,omitempty"`
	Amzn_Title           Amzn_Title           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Title,omitempty" json:"Title,omitempty"`
	XMLName              xml.Name             `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemAttributes,omitempty" json:"ItemAttributes,omitempty"`
}

type Amzn_ItemLink struct {
	Amzn_Description Amzn_Description `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Description,omitempty" json:"Description,omitempty"`
	Amzn_URL         Amzn_URL         `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	XMLName          xml.Name         `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemLink,omitempty" json:"ItemLink,omitempty"`
}

type Amzn_ItemLinks struct {
	Amzn_ItemLink []Amzn_ItemLink `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemLink,omitempty" json:"ItemLink,omitempty"`
	XMLName       xml.Name        `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemLinks,omitempty" json:"ItemLinks,omitempty"`
}

type Amzn_ItemPage struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemPage,omitempty" json:"ItemPage,omitempty"`
}

type Amzn_ItemSearchRequest struct {
	Amzn_ItemPage      Amzn_ItemPage        `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemPage,omitempty" json:"ItemPage,omitempty"`
	Amzn_Keywords      Amzn_Keywords        `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Keywords,omitempty" json:"Keywords,omitempty"`
	Amzn_ResponseGroup []Amzn_ResponseGroup `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ResponseGroup,omitempty" json:"ResponseGroup,omitempty"`
	Amzn_SearchIndex   Amzn_SearchIndex     `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SearchIndex,omitempty" json:"SearchIndex,omitempty"`
	XMLName            xml.Name             `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemSearchRequest,omitempty" json:"ItemSearchRequest,omitempty"`
}

type Amzn_ItemSearchResponse struct {
	Attr_xmlns            string                `xml:" xmlns,attr"  json:",omitempty"`
	Amzn_Items            Amzn_Items            `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Items,omitempty" json:"Items,omitempty"`
	Amzn_OperationRequest Amzn_OperationRequest `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 OperationRequest,omitempty" json:"OperationRequest,omitempty"`
	XMLName               xml.Name              `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemSearchResponse,omitempty" json:"ItemSearchResponse,omitempty"`
}

type Amzn_Items struct {
	Amzn_Item                 Amzn_Item                 `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Item,omitempty" json:"Item,omitempty"`
	Amzn_MoreSearchResultsUrl Amzn_MoreSearchResultsUrl `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 MoreSearchResultsUrl,omitempty" json:"MoreSearchResultsUrl,omitempty"`
	Amzn_Request              Amzn_Request              `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Request,omitempty" json:"Request,omitempty"`
	Amzn_TotalPages           Amzn_TotalPages           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 TotalPages,omitempty" json:"TotalPages,omitempty"`
	Amzn_TotalResults         Amzn_TotalResults         `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 TotalResults,omitempty" json:"TotalResults,omitempty"`
	XMLName                   xml.Name                  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Items,omitempty" json:"Items,omitempty"`
}

type Amzn_Keywords struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Keywords,omitempty" json:"Keywords,omitempty"`
}

type Amzn_Label struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Label,omitempty" json:"Label,omitempty"`
}

type Amzn_Language struct {
	Amzn_Name Amzn_Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Name,omitempty" json:"Name,omitempty"`
	Amzn_Type Amzn_Type `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Type,omitempty" json:"Type,omitempty"`
	XMLName   xml.Name  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Language,omitempty" json:"Language,omitempty"`
}

type Amzn_Languages struct {
	Amzn_Language Amzn_Language `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Language,omitempty" json:"Language,omitempty"`
	XMLName       xml.Name      `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Languages,omitempty" json:"Languages,omitempty"`
}

type Amzn_LargeImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 LargeImage,omitempty" json:"LargeImage,omitempty"`
}

type Amzn_Manufacturer struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Manufacturer,omitempty" json:"Manufacturer,omitempty"`
}

type Amzn_MediumImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 MediumImage,omitempty" json:"MediumImage,omitempty"`
}

type Amzn_Message struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Message,omitempty" json:"Message,omitempty"`
}

type Amzn_MoreSearchResultsUrl struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 MoreSearchResultsUrl,omitempty" json:"MoreSearchResultsUrl,omitempty"`
}

type Amzn_Name struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Name,omitempty" json:"Name,omitempty"`
}

type Amzn_NumberOfPages struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 NumberOfPages,omitempty" json:"NumberOfPages,omitempty"`
}

type Amzn_OperationRequest struct {
	Amzn_Arguments             Amzn_Arguments             `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Arguments,omitempty" json:"Arguments,omitempty"`
	Amzn_HTTPHeaders           Amzn_HTTPHeaders           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 HTTPHeaders,omitempty" json:"HTTPHeaders,omitempty"`
	Amzn_RequestId             Amzn_RequestId             `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 RequestId,omitempty" json:"RequestId,omitempty"`
	Amzn_RequestProcessingTime Amzn_RequestProcessingTime `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 RequestProcessingTime,omitempty" json:"RequestProcessingTime,omitempty"`
	XMLName                    xml.Name                   `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 OperationRequest,omitempty" json:"OperationRequest,omitempty"`
}

type Amzn_ProductGroup struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ProductGroup,omitempty" json:"ProductGroup,omitempty"`
}

type Amzn_ProductTypeName struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ProductTypeName,omitempty" json:"ProductTypeName,omitempty"`
}

type Amzn_PublicationDate struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 PublicationDate,omitempty" json:"PublicationDate,omitempty"`
}

type Amzn_Publisher struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Publisher,omitempty" json:"Publisher,omitempty"`
}

type Amzn_ReleaseDate struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ReleaseDate,omitempty" json:"ReleaseDate,omitempty"`
}

type Amzn_Request struct {
	Amzn_Errors            Amzn_Errors            `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Errors,omitempty" json:"Errors,omitempty"`
	Amzn_IsValid           Amzn_IsValid           `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 IsValid,omitempty" json:"IsValid,omitempty"`
	Amzn_ItemSearchRequest Amzn_ItemSearchRequest `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ItemSearchRequest,omitempty" json:"ItemSearchRequest,omitempty"`
	XMLName                xml.Name               `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Request,omitempty" json:"Request,omitempty"`
}

type Amzn_RequestId struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 RequestId,omitempty" json:"RequestId,omitempty"`
}

type Amzn_RequestProcessingTime struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 RequestProcessingTime,omitempty" json:"RequestProcessingTime,omitempty"`
}

type Amzn_ResponseGroup struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ResponseGroup,omitempty" json:"ResponseGroup,omitempty"`
}

type Amzn_SearchIndex struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SearchIndex,omitempty" json:"SearchIndex,omitempty"`
}

type Amzn_SmallImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SmallImage,omitempty" json:"SmallImage,omitempty"`
}

type Amzn_Source struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Source,omitempty" json:"Source,omitempty"`
}

type Amzn_Studio struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Studio,omitempty" json:"Studio,omitempty"`
}

type Amzn_SwatchImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 SwatchImage,omitempty" json:"SwatchImage,omitempty"`
}

type Amzn_ThumbnailImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 ThumbnailImage,omitempty" json:"ThumbnailImage,omitempty"`
}

type Amzn_TinyImage struct {
	Amzn_Height Amzn_Height `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Height,omitempty" json:"Height,omitempty"`
	Amzn_URL    Amzn_URL    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
	Amzn_Width  Amzn_Width  `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
	XMLName     xml.Name    `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 TinyImage,omitempty" json:"TinyImage,omitempty"`
}

type Amzn_Title struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Title,omitempty" json:"Title,omitempty"`
}

type Amzn_TotalPages struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 TotalPages,omitempty" json:"TotalPages,omitempty"`
}

type Amzn_TotalResults struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 TotalResults,omitempty" json:"TotalResults,omitempty"`
}

type Amzn_Type struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Type,omitempty" json:"Type,omitempty"`
}

type Amzn_URL struct {
	Text    string   `xml:",chardata" json:",omitempty"`
	XMLName xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 URL,omitempty" json:"URL,omitempty"`
}

type Amzn_Width struct {
	Attr_Units string   `xml:" Units,attr"  json:",omitempty"`
	Text       string   `xml:",chardata" json:",omitempty"`
	XMLName    xml.Name `xml:"http://webservices.amazon.com/AWSECommerceService/2013-08-01 Width,omitempty" json:"Width,omitempty"`
}
