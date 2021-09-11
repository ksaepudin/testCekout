package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/graphql-go/graphql"
)

type Product struct {
	SKU   string
	Name  string
	Price float32
	Qty   int
}
type Bonus struct {
	SKU   string
	Name  string
	Price float32
	Qty   int
}

type Basket struct {
	Products []Product
	Bonus    []Bonus
	Total    float32
}

var productType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "product",
		Fields: graphql.Fields{
			"sku": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"price": &graphql.Field{
				Type: graphql.Float,
			},
			"qty": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)
var bonusType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "bonus",
		Fields: graphql.Fields{
			"sku": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"price": &graphql.Field{
				Type: graphql.Float,
			},
			"qty": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

var basketType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "basket",
		Fields: graphql.Fields{
			"products": &graphql.Field{
				Type: graphql.NewList(productType),
			},
			"bonus": &graphql.Field{
				Type: graphql.NewList(bonusType),
			},
			"total": &graphql.Field{
				Type: graphql.Float,
			},
		},
	},
)

type reqBody struct {
	Query string `json:"query"`
}

func main() {
	http.HandleFunc("/", grqHandler)
	http.ListenAndServe(":8080", nil)
}

func grqHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "No Query Data", 400)
	}

	var rBody reqBody

	err := json.NewDecoder(r.Body).Decode(&rBody)
	if err != nil {
		http.Error(w, "Error parsing JSON request body", 400)
	}

	fmt.Fprintf(w, "%s", getData(rBody.Query))
}

func getData(query string) string {
	resultDataProduct := getDataJson
	params := graphql.Params{Schema: createSchema(resultDataProduct()), RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		fmt.Printf("failed to execute graphql operation, errors: %+v", r.Errors)
	}

	rJson, _ := json.Marshal(r)

	return fmt.Sprintf("%s", rJson)
}

func getDataJson() []Product {
	jfiles, err := os.Open("data.json")
	if err != nil {
		fmt.Printf("failed to open json file, error: %v", err)
	}
	jsonData, _ := ioutil.ReadAll(jfiles)
	defer jfiles.Close()
	var datProduct []Product
	err = json.Unmarshal(jsonData, &datProduct)
	if err != nil {
		fmt.Printf("failed to open json file, error: %v", err)
	}
	return datProduct
}

func createSchema(datProduct []Product) graphql.Schema {

	fields := graphql.Fields{
		"basket": &graphql.Field{
			Type:        basketType,
			Description: "Get Result Data Product",
			Args: graphql.FieldConfigArgument{
				"sku": &graphql.ArgumentConfig{
					Type: graphql.NewList(graphql.NewNonNull(graphql.String)),
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				type groupIDs struct {
					SKU string
					Qty int
				}

				var groupItem []groupIDs
				for _, v := range datProduct {
					var itemid int
					for _, gid := range p.Args["sku"].([]interface{}) {
						if v.SKU == gid {
							itemid = itemid + 1
						}
					}
					groupItem = append(groupItem, groupIDs{SKU: v.SKU, Qty: itemid})
				}

				var datBasket Basket

				var totalPrice, totalPriceAll float32
				// var countItem = 0
				for _, final := range groupItem {
					for _, v := range datProduct {
						if v.SKU == final.SKU && final.Qty != 0 {

							if v.SKU == "43N23P" { //macbook
								for _, x := range datProduct {
									if v.SKU == final.SKU {
										if x.SKU == "234234" {
											datBonus := Bonus{
												SKU:   x.SKU,
												Name:  x.Name,
												Price: 0,
												Qty:   1,
											}
											datBasket.Bonus = append(datBasket.Bonus, datBonus)
										}
									}
								}
								totalPrice = totalPrice + v.Price
							} else if v.SKU == "120P90" && final.Qty > 3 { //GoogleHome
								totalPriceItem := v.Price * float32(final.Qty-(final.Qty/3))
								v.Price = totalPriceItem
							} else if v.SKU == "120P90" && final.Qty < 3 {
								v.Price = v.Price * float32(final.Qty)
							}

							if v.SKU == "A304SD" && final.Qty == 3 { //alexa
								totalPriceItem := v.Price * float32(final.Qty)
								disItem := (v.Price * float32(final.Qty)) * 0.1
								v.Price = totalPriceItem - disItem
							} else if v.SKU == "A304SD" && final.Qty < 2 {
								v.Price = v.Price * float32(final.Qty)
							}

							v.Qty = final.Qty

							totalPriceAll = totalPriceAll + v.Price
							datBasket.Products = append(datBasket.Products, v)

						}
					}
					datBasket.Total = totalPriceAll
				}
				return datBasket, nil
			},
		},
	}

	rootquery := graphql.ObjectConfig{Name: "TestRySytem", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootquery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		fmt.Printf("failed to create new schema, error: %v", err)
	}

	return schema
}
