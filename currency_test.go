package money

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
)

func TestCurrency_Get(t *testing.T) {
	tcs := []struct {
		code     string
		expected string
	}{
		{EUR, "EUR"},
		{"EUR", "EUR"},
		{"Eur", "EUR"},
	}

	for _, tc := range tcs {
		c := newCurrency(tc.code).get()

		if c.Code != tc.expected {
			t.Errorf("Expected %s got %s", tc.expected, c.Code)
		}
	}
}

func TestCurrency_Get1(t *testing.T) {
	code := "RANDOM"
	c := newCurrency(code).get()

	if c.Grapheme != code {
		t.Errorf("Expected %s got %s", c.Grapheme, code)
	}
}

func TestCurrency_Equals(t *testing.T) {
	tcs := []struct {
		code  string
		other string
	}{
		{EUR, "EUR"},
		{"EUR", "EUR"},
		{"Eur", "EUR"},
		{"usd", "USD"},
	}

	for _, tc := range tcs {
		c := newCurrency(tc.code).get()
		oc := newCurrency(tc.other).get()

		if !c.equals(oc) {
			t.Errorf("Expected that %v is not equal %v", c, oc)
		}
	}
}

func TestCurrency_AddCurrency(t *testing.T) {
	tcs := []struct {
		code     string
		template string
	}{
		{"GOLD", "1$"},
	}

	for _, tc := range tcs {
		AddCurrency(tc.code, "", tc.template, "", "", 0)
		c := newCurrency(tc.code).get()

		if c.Template != tc.template {
			t.Errorf("Expected currency template %v got %v", tc.template, c.Template)
		}
	}
}

func TestCurrency_GetCurrency(t *testing.T) {
	code := "KLINGONDOLLAR"
	desired := Currency{Decimal: ".", Thousand: ",", Code: code, Fraction: 2, Grapheme: "$", Template: "$1"}
	AddCurrency(desired.Code, desired.Grapheme, desired.Template, desired.Decimal, desired.Thousand, desired.Fraction)
	currency := GetCurrency(code)
	if !reflect.DeepEqual(currency, &desired) {
		t.Errorf("Currencies do not match %+v got %+v", desired, currency)
	}
}

func TestCurrency_GetNonExistingCurrency(t *testing.T) {
	currency := GetCurrency("I*am*Not*a*Currency")
	if currency != nil {
		t.Errorf("Unexpected currency returned %+v", currency)
	}
}

func TestGetCurrencyByNumericCode(t *testing.T) {
	type args struct {
		code string
	}
	tests := []struct {
		name string
		args args
		want *Currency
	}{
		{
			"happy-currency-find",
			args{code: "986"},
			&Currency{Decimal: ",", Thousand: ".", Code: BRL, Fraction: 2, NumericCode: "986", Grapheme: "R$", Template: "$1"},
		},
		{
			"happy-currency-not-found",
			args{code: "1111"},
			nil,
		},
		{
			"happy-currency-empty",
			args{code: ""},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CurrencyByNumericCode(tt.args.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCurrencyByNumericCode() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCurrencyCodeByXmlIsoFile(t *testing.T) {
	xmlIsoFile := getXmlIsoFile()
	for _, currency := range xmlIsoFile.Table.Items {
		if currency.CurrencyName != "No universal currency" {

			t.Run(fmt.Sprintf("testing %s", currency.NumericCurrencyCode), func(t *testing.T) {
				c := CurrencyByNumericCode(currency.NumericCurrencyCode)
				if !reflect.DeepEqual(c.NumericCode, currency.NumericCurrencyCode) {
					t.Errorf("GetCurrencyByNumericCode() got = %v, want %v", c.NumericCode, currency.NumericCurrencyCode)
				}
			})

			t.Run(fmt.Sprintf("testing %s", currency.CurrencyCode), func(t *testing.T) {
				c := GetCurrency(currency.CurrencyCode)
				if !reflect.DeepEqual(c.Code, currency.CurrencyCode) {
					t.Errorf("GetCurrency() got = %v, want %v", c.Code, currency.CurrencyCode)
				}
			})

		}
	}
}

const (
	urlIsoCurrency = "https://www.six-group.com/dam/download/financial-information/data-center/iso-currrency/amendments/lists/list_one.xml"
	downloadedFile = "/tmp/downloaded_file.xml"
	localFile      = "../../../../scripts/currencies/list_one.xml"
)

type (
	XmlIsoFile struct {
		Iso        xml.Name    `xml:"ISO_4217"`
		Publishing string      `xml:"Pblshd,attr"`
		Table      IsoCurrency `xml:"CcyTbl"`
	}
	IsoCurrency struct {
		Items []Item `xml:"CcyNtry"`
	}

	Item struct {
		Country             string `xml:"CtryNm"`
		CurrencyName        string `xml:"CcyNm"`
		CurrencyCode        string `xml:"Ccy"`
		NumericCurrencyCode string `xml:"CcyNbr"`
		DecimalUnits        string `xml:"CcyMnrUnts"`
	}
)

var xmlIsoFile *XmlIsoFile

func ParseIsoXml() *XmlIsoFile {
	fileUrl := urlIsoCurrency
	err := DownloadFile(downloadedFile, fileUrl)

	var xmlFile *os.File
	if err != nil {
		fmt.Println("\tUsing local xml file")
		xmlFile, err = os.Open(localFile)
	} else {
		fmt.Println("\tUsing downloads xml file")
		xmlFile, err = os.Open(downloadedFile)
	}
	// if os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println("\tSuccessfully opened list_one.xml")
	// defer the closing of xmlFile so that we can parse it.
	defer xmlFile.Close()
	byteValue, _ := ioutil.ReadAll(xmlFile)
	// Unmarshal takes a []byte and fills the rss struct with the values found in the xmlFile
	xml.Unmarshal(byteValue, &xmlIsoFile)
	fmt.Println("Rss version: " + xmlIsoFile.Publishing)

	return xmlIsoFile
}

func getXmlIsoFile() *XmlIsoFile {
	if xmlIsoFile == nil {
		xmlIsoFile = ParseIsoXml()
	}
	return xmlIsoFile
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
