package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func Perform(args Arguments, writer io.Writer) error {
	oper, ok := args["operation"]
	if !ok || len(oper) == 0 {
		return fmt.Errorf("specify -operation argument")
	}

	filename, ok := args["fileName"]
	if !ok || len(filename) == 0 {
		return fmt.Errorf("specify -fileName argument")
	}

	switch oper {
	case "add":
		return add(filename, args["item"], writer)
	case "list":
		return list(filename, writer)
	case "findById":
		return findById(filename, args["id"], writer)
	case "remove":
		return remove(filename, args["id"], writer)
	}

	return fmt.Errorf("Invalid operation: %s ", oper)
}

func parseArgs() Arguments {
	names := []string{"id", "item", "operation", "fileName"}
	values := make([]string, len(names))

	for i := range names {
		flag.StringVar(&values[i], names[i], "", "flag "+names[i])
	}

	flag.Parse()

	var args Arguments = make(map[string]string)

	for i := range names {
		if len(values[i]) > 0 {
			args[names[i]] = values[i]
		}
	}

	return args
}

// io
func readFile(fileName string) ([]byte, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func writeFile(filename string, body []byte) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(body)
	return err
}

func readItems(filename string) ([]User, error) {
	blob, err := readFile(filename)
	if err != nil {
		return nil, err
	}

	if len(blob) == 0 {
		return make([]User, 0), nil
	}

	var items []User
	err = json.Unmarshal(blob, &items)
	if err != nil {
		return nil, err
	}

	return items, nil
}

func writeItems(filename string, users []User) error {
	blob, err := json.Marshal(users)
	if err != nil {
		return err
	}

	return writeFile(filename, blob)
}

// operations
func add(filename string, itemData string, writer io.Writer) error {
	if len(itemData) == 0 {
		return fmt.Errorf("specify -item argument")
	}

	users, err := readItems(filename)
	if err != nil {
		return err
	}

	var user User
	err = json.Unmarshal([]byte(itemData), &user)
	if err != nil {
		return err
	}

	for _, v := range users {
		if v.Id == user.Id {
			writer.Write([]byte(fmt.Sprintf("id %s is already taken", user.Id)))
			return nil
		}
	}

	users = append(users, user)
	return writeItems(filename, users)
}

func list(filename string, writer io.Writer) error {
	body, err := readFile(filename)
	if err != nil {
		return err
	}
	writer.Write(body)
	return nil
}

func findById(filename string, id string, writer io.Writer) error {
	if len(id) == 0 {
		return fmt.Errorf("specify -id argument")
	}

	items, err := readItems(filename)
	if err != nil {
		return err
	}

	for _, v := range items {
		if v.Id == id {
			blob, err := json.Marshal(v)
			if err != nil {
				return err
			}
			_, err = writer.Write(blob)
			return err
		}
	}

	return nil
}

func remove(filename string, id string, writer io.Writer) error {
	if len(id) == 0 {
		return fmt.Errorf("specify -id argument")
	}

	items, err := readItems(filename)
	if err != nil {
		return err
	}

	//find
	for i, v := range items {
		if v.Id == id {
			items = append(items[:i], items[i+1:]...)
			return writeItems(filename, items)
		}
	}

	//ok, not found
	writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
	return nil
}
