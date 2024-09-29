// Compiled Program
// Language: Go
// HTML used as template for the UI
// Running this program starts a local web server and will automatically open the webpage
// To run the program just run ProbCalc.exe
package main

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// type Data struct{
// 	Dice   int
// 	Sides  int
// 	Sum    int
// 	Answer string
// }

func main() {
	// Parse the template file
	server := &http.Server{Addr: ":8080"}
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}

	// Handle the form submission and response
	http.HandleFunc("/probability", func(w http.ResponseWriter, r *http.Request) {
		// Create a struct to hold the form data and answer
		type Data struct {
			Dice    int
			Sides   int
			Sum     int
			Answer  string
			Subsets []int
			FormResult string
		}

		var data Data
		sum_result := ""
		indiv_result := ""
		// Handle POST request to process the form data
		if r.Method == http.MethodPost {
			// Retrieve form values and convert them to integers
			dice, _ := strconv.Atoi(r.FormValue("dice"))
			sides, _ := strconv.Atoi(r.FormValue("sides"))
			formresult:= strings.TrimSpace(r.FormValue("answer"))
			formresultlines := strings.Split(formresult, "\n")
			fmt.Printf("form result value: %s \n",formresult)
			// for i, line := range formresultlines {
			// 	fmt.Printf("Line %d: %s\n", i+1 , line)
			// }
			results := [][]int{}

			//get all possible combinations
			rollDice(dice, sides, []int{}, &results)

			// Check which button was clicked by looking at the name/value of the submit button
			if r.FormValue("action") == "Sum Probability" {
				// Retrieve sum values
				sum, _ := strconv.Atoi(r.FormValue("sum"))
				data.Sum = sum
				// Call the probability function (placeholder logic)
				sum_result = fmt.Sprint(probability(sumRes(results, sum), len(results)))
				answer := fmt.Sprintf("Probability of getting the total: %s %%", sum_result)
				data.Answer = answer + " <br>\n" + formresultlines[1]
			} else if r.FormValue("action") == "Combination Probability" {
				// Retrieve subset values (the dynamically created inputs)
				subsetValues := r.Form["subset[]"]
				var subsets []int
				for _, val := range subsetValues {
					if intVal, err := strconv.Atoi(val); err == nil {
						subsets = append(subsets, intVal)
					}
				}

				// Store the subsets in the data and create a response
				data.Subsets = subsets
				indiv_result = fmt.Sprint(probability(indivRes(&subsets, &results), len(results)))
				answer := fmt.Sprintf("Probability for getting specific combination: %s%%", indiv_result)
				data.Answer = formresultlines[0] + "<br>\n" + answer
			}

			// Set the data struct with values from the form and answer
			data.Dice = dice
			data.Sides = sides
			// Render only the answer part (used by HTMX for partial update)
			tmpl.ExecuteTemplate(w, "answer", data)
			return
		}

		// For GET request, show the form with no answer initially
		data = Data{Answer: "Probability of getting the total: _ \n<br>Probability for getting specific combination: _ "}
		tmpl.ExecuteTemplate(w, "index.html", data)
	})

	// Start the server
	
		// Handle server shutdown
		http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Server is shutting down...")
			go func() {
				time.Sleep(1 * time.Second)
				if err := server.Shutdown(context.Background()); err != nil {
					fmt.Println("Server Shutdown:", err)
				}
			}()
		})
	
		// Start the server
		fmt.Println("Server is running on http://localhost:8080")
		openBrowser("http://localhost:8080/probability")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
}

//PROBABILITY FORMULA
func probability(desResult int, posResult int) float32 {
	result := (float32(desResult) / float32(posResult)) * 100
	return result
}

// ALL OUTCOME
func rollDice(dn, ds int, result []int, results *[][]int) {
	// Base case: If we have rolled all the dice, append the result
	if dn == 0 {
		*results = append(*results, append([]int{}, result...))
		return
	}

	// Try all sides of the dice for the current roll
	for i := 1; i <= ds; i++ {
		rollDice(dn-1, ds, append(result, i), results)
	}
}

//INDIVIDUAL OUTCOME
func indivRes(desRes *[]int, allRes *[][]int) int {

	// Array to store results that contain the subset x
	validResults := [][]int{}

	// Check which combinations contain subset x and store them
	for _, result := range *allRes{
		if containsSubset(result, *desRes) {
			validResults = append(validResults, result)
		}
	}
	//Printing results
	// fmt.Printf("Combinations that contain the subset %v:\n", desRes)
	// for _, result := range validResults {
	// 	fmt.Println(result)
	// }
	return len(validResults)
}

// TOTAL OUTCOME
func sumRes(results [][]int, target int) int {
	validCombinations := [][]int{}

	// Iterate over all the results and calculate the sum
	for _, result := range results {
		sum := 0
		for _, num := range result {
			sum += num
		}
		// If the sum equals the target, add the result to the valid combinations
		if sum == target {
			validCombinations = append(validCombinations, result)
		}
	}
	// Printing Results
	// fmt.Printf("Combinations that totals the target %d:\n", target)
	// for _, result := range validCombinations {
	// 	fmt.Println(result)
	// }
	return len(validCombinations)
}



// containsSubset checks if result contains all elements of the subset
func containsSubset(result, subset []int) bool {
	// Create a frequency map for both result and subset
	resultCount := make(map[int]int)
	for _, val := range result {
		resultCount[val]++
	}

	for _, val := range subset {
		// If an element in subset is missing or occurs less often in result, return false
		if resultCount[val] == 0 {
			return false
		}
		resultCount[val]--
	}

	return true
}

//function for opening browser when running the application
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin": // macOS
		err = exec.Command("open", url).Start()
	}
	if err != nil {
		fmt.Println("Error opening browser:", err)
	}
}