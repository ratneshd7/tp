package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Item contains information about one Item
type Item struct {
	ItemIndex int64  `json:"itemindex"`
	ItemName  string `json:"itemname"`
}

var outputFileName string = "dummyFile.json"

// Items List
var Items []Item

func autogen() []Item {
	var items []Item
	var item Item
	var i int64
	for i = 1; i < 20; i++ {
		item.ItemIndex = i
		item.ItemName = fmt.Sprintf("name%v", i)
		items = append(items, item)
	}
	return items
}

// Write to File.
func writeToFile(d []byte) {
	err := ioutil.WriteFile(outputFileName, d, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
}

// Read Data From File.
func readFile() []byte {
	data, err := ioutil.ReadFile(outputFileName)
	if err != nil {
		log.Fatal("Not Able to Read File")
	}
	return data
}

func encodeByteToJSON(data []byte) {
	var items []Item
	err := json.Unmarshal(data, &items)
	if err != nil {
		log.Fatal(err.Error())
	}

	// Assign to Global Variable.
	Items = items
}

// ItemType Hold New Object Of Item
var ItemType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Item",
		Fields: graphql.Fields{
			"itemindex": &graphql.Field{
				Type: graphql.Int,
			},
			"itemname": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var queryType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			/* Get (read) single Item by id
			   http://localhost:8080/Item?query={Item(itemindex:1){itemname}}
			*/
			"Item": &graphql.Field{
				Type:        ItemType,
				Description: "Get Item by id",
				Args: graphql.FieldConfigArgument{
					"itemindex": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["itemindex"].(int)
					if ok {
						// Read From File
						read()
						// Find Item
						for _, Item := range Items {
							if int(Item.ItemIndex) == id {
								return Item, nil
							}
						}
					}
					return nil, nil
				},
			},
			/* Get (read) Item list
			   http://localhost:8080/Item?query={list{itemindex, itemname}}
			*/
			"list": &graphql.Field{
				Type:        graphql.NewList(ItemType),
				Description: "Get Item list",
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					// Read From File
					read()
					return Items, nil
				},
			},
		},
	})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Mutation",
	Fields: graphql.Fields{
		/* Create new Item item
		http://localhost:8080/Item?query=mutation+_{create(itemindex:20,itemname:"Inca Kola"){itemname}}
		*/
		"create": &graphql.Field{
			Type:        ItemType,
			Description: "Create new Item",
			Args: graphql.FieldConfigArgument{
				"itemindex": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"itemname": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				// Read From File.
				read()
				Item := Item{
					ItemIndex: int64(params.Args["itemindex"].(int)),
					ItemName:  params.Args["itemname"].(string),
				}
				Items = append(Items, Item)
				// Write To File.
				write()
				return Item, nil
			},
		},

		/* Update Item by id
		http://localhost:8080/Item?query=mutation+_{update(itemindex:1,itemname:"Change Value"){itemindex, itemname}}
		*/
		"update": &graphql.Field{
			Type:        ItemType,
			Description: "Update Item by id",
			Args: graphql.FieldConfigArgument{
				"itemindex": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
				"itemname": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, _ := params.Args["itemindex"].(int)
				name, nameOk := params.Args["itemname"].(string)
				Item := Item{}
				// Read From File.
				read()
				for i, p := range Items {
					if int64(id) == p.ItemIndex {
						if nameOk {
							Items[i].ItemName = name
						}
						Item = Items[i]
						break
					}
				}
				// Write To File.
				write()
				return Item, nil
			},
		},

		/* Delete Item by id
		   http://localhost:8080/Item?query=mutation+_{delete(itemindex:1){itemindex, itemname}}
		*/
		"delete": &graphql.Field{
			Type:        ItemType,
			Description: "Delete Item by id",
			Args: graphql.FieldConfigArgument{
				"itemindex": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.Int),
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				id, _ := params.Args["itemindex"].(int)
				Item := Item{}
				// Read List From File.
				read()
				for i, p := range Items {
					if int64(id) == p.ItemIndex {
						Item = Items[i]
						// Remove from Item list
						Items = append(Items[:i], Items[i+1:]...)
					}
				}
				// Write To File.
				write()
				return Item, nil
			},
		},
	},
})

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	},
)

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}

func read() {
	data := readFile()
	encodeByteToJSON(data)
}

func write() {
	d, _ := json.Marshal(&Items)
	writeToFile(d)
}

func main() {
	http.HandleFunc("/Item", func(w http.ResponseWriter, r *http.Request) {
		result := executeQuery(r.URL.Query().Get("query"), schema)
		json.NewEncoder(w).Encode(result)
	})
	fmt.Println("Server is running on port 8080")
	http.ListenAndServe(":8080", nil)
}
