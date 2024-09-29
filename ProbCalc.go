// Compiled Program
// Language: Go
// HTML used as template for the UI
// Running this program starts a local web server and will automatically open the webpage
// To run the program just run ProbCalc.exe
package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"
)


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
			//fmt.Printf("form result value: %s \n",formresult)

			// Check which button was clicked by looking at the name/value of the submit button
			if r.FormValue("action") == "Sum Probability" {
				// Retrieve sum values
				sum, _ := strconv.Atoi(r.FormValue("sum"))
				data.Sum = sum
				start := time.Now()
				fmt.Println("Stopwatch started...")
				sum_result = formatNumber(sumProbability(dice, sides, sum))
				answer := fmt.Sprintf("Probability of getting the total: %s %%", sum_result)
				data.Answer = answer + " <br>\n" + formresultlines[1]
				elapsed := time.Since(start)
				fmt.Printf("Elapsed time: %s\n", elapsed)
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
				start := time.Now()
				fmt.Println("Stopwatch started...")
				indiv_result = formatNumber(indivProbability(dice, sides, subsets))
				answer := fmt.Sprintf("Probability for getting specific combination: %s%%", indiv_result)
				elapsed := time.Since(start)
				fmt.Printf("Elapsed time: %s\n", elapsed)
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
func probability(desResult int, posResult int) float64 {
	result := (float64(desResult) / float64(posResult)) * 100
	return result
}
//SUM PROBABILITY CALCULATOR
func sumProbability(dn, ds, sum int) float64 {
	// Helper function to calculate all possible combinations recursively.
	var countCombinations func(dn, sum, sides int) int
	countCombinations = func(dn, sum, sides int) int {
		if dn == 0 {
			if sum == 0 {
				return 1
			}
			return 0
		}
		count := 0
		for i := 1; i <= sides; i++ {
			count += countCombinations(dn-1, sum-i, sides)
		}
		return count
	}
	// Calculate the number of possible rolls resulting in the sum.
	desRes := countCombinations(dn, sum, ds)
	// Calculate the number of all possible rolls.
	totalPossibleRolls := int(math.Pow(float64(ds), float64(dn)))
	// Call the provided probability function.
	return probability(desRes, totalPossibleRolls)
}


//COMBINATION PROBABILITY CALCULATOR
func indivProbability(dn, ds int, subset []int) float64 {
	count := 0
	// Helper function to roll the dice and check the subset.
	var rollDice func(dn int, result []int)
	rollDice = func(dn int, result []int) {
		// Base case: If we have rolled all the dice, check the result.
		if dn == 0 {
			if containsSubset(result, subset) {
				count++
			}
			return
		}
		// Try all sides of the dice for the current roll.
		for i := 1; i <= ds; i++ {
			rollDice(dn-1, append(result, i))
		}
	}
	// Total possible rolls.
	totalCount := int(math.Pow(float64(ds), float64(dn)))
	// Start rolling the dice.
	rollDice(dn, []int{})
	// Calculate and return the probability.
	return probability(count, totalCount)
}

// containsSubset checks if result contains all elements of the subset.
func containsSubset(result []int, subset []int) bool {
	// Create a frequency map for both result and subset.
	resultCount := make(map[int]int)

	for _, val := range result {
		resultCount[val]++
	}

	for _, val := range subset {
		// If an element in subset is missing or occurs less often in result, return false.
		if resultCount[val] == 0 {
			return false
		}
		resultCount[val]--
	}

	return true
}

func formatNumber(value float64) string {
	if value == 0 {
		return "0" // Display zero as "0"
	}
	if value < 0.01 && value > -0.01 {
		return fmt.Sprintf("%.2e", value) // Use scientific notation for very small values
	}
	return fmt.Sprintf("%.2f", value) // Use standard decimal format otherwise
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