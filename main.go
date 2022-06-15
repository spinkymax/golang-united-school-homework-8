package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Users []User

func main() {
	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(err)
	}
}

func parseArgs() Arguments {
	fileName := flag.String("fileName", "", "filename")
	operation := flag.String("operation", "", "operation")
	item := flag.String("item", "", "item")
	flag.Parse()

	return Arguments{
		"operation": *operation,
		"item":      *item,
		"fileName":  *fileName}
}

func Perform(args Arguments, writer io.Writer) error {

	operation := args["operation"]
	if operation == "" {
		return errors.New("-operation flag has to be specified")
	}

	fileName := args["fileName"]
	if fileName == "" {
		return errors.New("-fileName flag has to be specified")
	}

	switch operation {
	case "add":
		item := args["item"]
		if item == "" {
			return errors.New("-item flag has to be specified")
		}
		return add(item, fileName, writer)

	case "list":
		{
			str := list(args["fileName"])
			_, err := writer.Write([]byte(str))
			if err != nil {
				return err
			}
		}

	case "findById":
		{
			if args["id"] == "" {
				return errors.New("-id flag has to be specified")
			}

			str := findById(args["id"], args["fileName"])
			_, err := writer.Write([]byte(str))
			if err != nil {
				return err
			}
		}

	case "remove":
		{
			if args["id"] == "" {
				return errors.New("-id flag has to be specified")
			}

			str := remove(args["id"], args["fileName"])
			_, err := writer.Write([]byte(str))
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Operation %s not allowed!", operation)
	}
	return nil
}

func list(fileName string) string {

	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return err.Error()
	}
	defer file.Close()
	data, _ := ioutil.ReadAll(file)
	return string(data)
}

func add(item, fileName string, writer io.Writer) error {

	var newUser User
	if err := json.Unmarshal([]byte(item), &newUser); err != nil {
		return fmt.Errorf("Sorry, but I can't unmarshal the new user JSON: %w", err)
	}

	var users []User
	bytes, _ := os.ReadFile(fileName)
	if err := json.Unmarshal(bytes, &users); err != nil {
		users = make([]User, 0, 1)
	} else {
		for _, user := range users {
			if user.Id == newUser.Id {
				writer.Write([]byte("Item with id" + user.Id + "already exists"))
				return nil
			}
		}
	}

	users = append(users, newUser)
	if err := save(users, fileName); err != nil {
		return fmt.Errorf("Sorry, but I can't save users: %v", err)
	}

	return nil
}

func save(users []User, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	bytes, err := json.Marshal(users)
	if err != nil {
		return err
	}

	file.Write(bytes)

	return nil
}
func findById(id, fileName string) string {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
	if err != nil {
		return "Sorry, but I can't open file"
	}
	defer file.Close()

	if stat, _ := file.Stat(); stat.Size() == 0 {
		return "File " + fileName + " is empty"
	}

	objects := []User{}
	data, err := ioutil.ReadAll(file)
	err = json.Unmarshal(data, &objects)
	if err != nil {
		return "Unmarshal from file error"
	}

	for _, objectJson := range objects {
		if objectJson.Id == id {
			result, _ := json.Marshal(objectJson)
			return string(result)
		}
	}

	return ""
}

func remove(id, fileName string) string {
	file, err := os.OpenFile(fileName, os.O_RDWR, 0644)
	if err != nil {
		return "Sorry, but I can't open it"
	}
	defer file.Close()

	if stat, _ := file.Stat(); stat.Size() == 0 {
		return "It " + fileName + " is empty"
	}

	objects := []User{}
	data, err := ioutil.ReadAll(file)
	err = json.Unmarshal(data, &objects)
	if err != nil {
		return ""
	}

	for i, objectJson := range objects {
		if objectJson.Id == id {
			objects = append(objects[:i], objects[i+1:]...)
			result, _ := json.Marshal(objects)
			file.Truncate(0)
			file.Seek(0, 0)		
			file.Write(result)

			return ""
		}
	}
	return "Item with " + id + " not found"
}

