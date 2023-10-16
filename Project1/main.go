package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func main() {
	// CLI args
	f, closeFile, err := openProcessingFile(os.Args...)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFile()

	// Load and parse processes
	processes, err := loadProcesses(f)
	if err != nil {
		log.Fatal(err)
	}

	// Define the time quantum (you can adjust this value as needed)
	var timeQuantum int64 = 2

	// First-come, first-serve scheduling
	FCFSSchedule(os.Stdout, "First-come, first-serve", processes)

	// SJF (preemptive) scheduling
	SJFSchedule(os.Stdout, "Shortest-job-first (preemptive)", processes)

	//SJF Priority Scheduling (preemptive)
	SJFPrioritySchedule(os.Stdout, "Priority", processes)

	RRSchedule(os.Stdout, "Round-robin", processes, timeQuantum)
}

func openProcessingFile(args ...string) (*os.File, func(), error) {
	if len(args) != 2 {
		return nil, nil, fmt.Errorf("%w: must give a scheduling file to process", ErrInvalidArgs)
	}
	// Read in CSV process CSV file
	f, err := os.Open(args[1])
	if err != nil {
		return nil, nil, fmt.Errorf("%v: error opening scheduling file", err)
	}
	closeFn := func() {
		if err := f.Close(); err != nil {
			log.Fatalf("%v: error closing scheduling file", err)
		}
	}

	return f, closeFn, nil
}

type (
	Process struct {
		ProcessID     int64
		ArrivalTime   int64
		BurstDuration int64
		Priority      int64
	}
	TimeSlice struct {
		PID   int64
		Start int64
		Stop  int64
	}
)

//region Schedulers

// FCFSSchedule outputs a schedule of processes in a GANTT chart and a table of timing given:
// • an output writer
// • a title for the chart
// • a slice of processes
func FCFSSchedule(w io.Writer, title string, processes []Process) {
	var (
		serviceTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  float64
		waitingTime     int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
	)
	for i := range processes {
		if processes[i].ArrivalTime > 0 {
			waitingTime = serviceTime - processes[i].ArrivalTime
		}
		totalWait += float64(waitingTime)

		start := waitingTime + processes[i].ArrivalTime

		turnaround := processes[i].BurstDuration + waitingTime
		totalTurnaround += float64(turnaround)

		completion := processes[i].BurstDuration + processes[i].ArrivalTime + waitingTime
		lastCompletion = float64(completion)

		schedule[i] = []string{
			fmt.Sprint(processes[i].ProcessID),
			fmt.Sprint(processes[i].Priority),
			fmt.Sprint(processes[i].BurstDuration),
			fmt.Sprint(processes[i].ArrivalTime),
			fmt.Sprint(waitingTime),
			fmt.Sprint(turnaround),
			fmt.Sprint(completion),
		}
		serviceTime += processes[i].BurstDuration

		gantt = append(gantt, TimeSlice{
			PID:   processes[i].ProcessID,
			Start: start,
			Stop:  serviceTime,
		})
	}

	count := float64(len(processes))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / lastCompletion

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

// Common scheduling function with priority criteria
func schedule(w io.Writer, title string, processes []Process, priority func(int64, int64, int64) bool) {
	var (
		currentTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
		readyQueue      = make([]Process, 0)
	)

	for len(readyQueue) > 0 || len(processes) > 0 {
		// Add arriving processes to the ready queue
		for len(processes) > 0 && processes[0].ArrivalTime <= currentTime {
			readyQueue = append(readyQueue, processes[0])
			processes = processes[1:]
		}

		if len(readyQueue) == 0 {
			currentTime++
			continue
		}

		// Find the process with the highest priority (and shortest remaining burst time)
		highestPriorityIndex := 0
		for i := 1; i < len(readyQueue); i++ {
			if priority(readyQueue[i].Priority, readyQueue[i].BurstDuration, readyQueue[highestPriorityIndex].BurstDuration) {
				highestPriorityIndex = i
			}
		}

		currentProcess := readyQueue[highestPriorityIndex]
		readyQueue = append(readyQueue[:highestPriorityIndex], readyQueue[highestPriorityIndex+1:]...)

		// Calculate waiting time for the selected process
		waitingTime := currentTime - currentProcess.ArrivalTime
		totalWait += float64(waitingTime)

		// Update the gantt chart
		gantt = append(gantt, TimeSlice{
			PID:   currentProcess.ProcessID,
			Start: currentTime,
			Stop:  currentTime + currentProcess.BurstDuration,
		})

		// Update the scheduling information
		schedule[currentProcess.ProcessID-1] = []string{
			fmt.Sprint(currentProcess.ProcessID),
			fmt.Sprint(currentProcess.Priority),
			fmt.Sprint(currentProcess.BurstDuration),
			fmt.Sprint(currentProcess.ArrivalTime),
			fmt.Sprint(waitingTime),
			"", // Turnaround time (to be calculated later)
			"", // Completion time (to be calculated later)
		}

		// Update current time
		currentTime += currentProcess.BurstDuration

		// Calculate turnaround time for the selected process
		turnaround := currentTime - currentProcess.ArrivalTime
		totalTurnaround += float64(turnaround)

		lastCompletion = currentTime

		// Update the schedule with the calculated turnaround and completion time
		schedule[currentProcess.ProcessID-1][5] = fmt.Sprint(turnaround)
		schedule[currentProcess.ProcessID-1][6] = fmt.Sprint(currentTime)
	}

	count := float64(len(schedule))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / float64(lastCompletion)

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

// Function to determine priority based on SJF criteria
func sjfPriorityCriteria(priority, burst1, burst2 int64) bool {
	// Consider priority first; if priorities are the same, choose the process with the shorter burst time.
	if priority == burst1 && priority == burst2 {
		return false
	}
	if priority == burst1 {
		return true
	}
	if priority == burst2 {
		return false
	}
	return burst1 < burst2
}

// SJFSchedule performs Shortest-Job-First (preemptive) scheduling
func SJFSchedule(w io.Writer, title string, processes []Process) {
	schedule(w, title, processes, sjfPriorityCriteria)
}

// Function to determine priority based on SJF Priority criteria
func sjfPriorityPriorityCriteria(priority, burst1, burst2 int64) bool {
	// Choose the process with the shorter burst time and, if they are the same, select the one with higher priority.
	if burst1 == burst2 {
		return priority > priority
	}
	return burst1 < burst2
}

// SJFPrioritySchedule performs Shortest-Job-First Priority (preemptive) scheduling
func SJFPrioritySchedule(w io.Writer, title string, processes []Process) {
	schedule(w, title, processes, sjfPriorityPriorityCriteria)
}

// RRSchedule performs Round-Robin (preemptive) scheduling
func RRSchedule(w io.Writer, title string, processes []Process, timeQuantum int64) {
	var (
		currentTime     int64
		totalWait       float64
		totalTurnaround float64
		lastCompletion  int64
		schedule        = make([][]string, len(processes))
		gantt           = make([]TimeSlice, 0)
		readyQueue      = make([]Process, 0)
	)

	for len(readyQueue) > 0 || len(processes) > 0 {
		// Add arriving processes to the ready queue
		for len(processes) > 0 && processes[0].ArrivalTime <= currentTime {
			readyQueue = append(readyQueue, processes[0])
			processes = processes[1:]
		}

		if len(readyQueue) == 0 {
			currentTime++
			continue
		}

		// Get the first process in the ready queue
		currentProcess := readyQueue[0]

		// Determine the time slice for this process (limited by time quantum)
		timeSlice := mini(currentProcess.BurstDuration, timeQuantum)

		// Update the gantt chart
		gantt = append(gantt, TimeSlice{
			PID:   currentProcess.ProcessID,
			Start: currentTime,
			Stop:  currentTime + timeSlice,
		})

		// Update the scheduling information
		schedule[currentProcess.ProcessID-1] = []string{
			fmt.Sprint(currentProcess.ProcessID),
			fmt.Sprint(currentProcess.Priority),
			fmt.Sprint(currentProcess.BurstDuration),
			fmt.Sprint(currentProcess.ArrivalTime),
			"", // Waiting time (to be calculated later)
			"", // Turnaround time (to be calculated later)
			"", // Completion time (to be calculated later)
		}

		// Update current time
		currentTime += timeSlice

		// Update the process's burst duration
		currentProcess.BurstDuration -= timeSlice

		// Move the current process to the end of the ready queue if it's not completed
		if currentProcess.BurstDuration > 0 {
			readyQueue = append(readyQueue[1:], currentProcess)
		} else {
			// Process has completed
			// Calculate waiting time, turnaround time, and completion time
			waitingTime := currentTime - currentProcess.ArrivalTime - currentProcess.BurstDuration
			totalWait += float64(waitingTime)

			turnaround := currentTime - currentProcess.ArrivalTime
			totalTurnaround += float64(turnaround)

			schedule[currentProcess.ProcessID-1][4] = fmt.Sprint(waitingTime)
			schedule[currentProcess.ProcessID-1][5] = fmt.Sprint(turnaround)
			schedule[currentProcess.ProcessID-1][6] = fmt.Sprint(currentTime)
		}

		lastCompletion = currentTime
	}

	count := float64(len(schedule))
	aveWait := totalWait / count
	aveTurnaround := totalTurnaround / count
	aveThroughput := count / float64(lastCompletion)

	outputTitle(w, title)
	outputGantt(w, gantt)
	outputSchedule(w, schedule, aveWait, aveTurnaround, aveThroughput)
}

// min returns the minimum of two integers
func mini(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

//endregion

//region Output helpers

func outputTitle(w io.Writer, title string) {
	_, _ = fmt.Fprintln(w, strings.Repeat("-", len(title)*2))
	_, _ = fmt.Fprintln(w, strings.Repeat(" ", len(title)/2), title)
	_, _ = fmt.Fprintln(w, strings.Repeat("-", len(title)*2))
}

func outputGantt(w io.Writer, gantt []TimeSlice) {
	_, _ = fmt.Fprintln(w, "Gantt schedule")
	_, _ = fmt.Fprint(w, "|")
	for i := range gantt {
		pid := fmt.Sprint(gantt[i].PID)
		padding := strings.Repeat(" ", (8-len(pid))/2)
		_, _ = fmt.Fprint(w, padding, pid, padding, "|")
	}
	_, _ = fmt.Fprintln(w)
	for i := range gantt {
		_, _ = fmt.Fprint(w, fmt.Sprint(gantt[i].Start), "\t")
		if len(gantt)-1 == i {
			_, _ = fmt.Fprint(w, fmt.Sprint(gantt[i].Stop))
		}
	}
	_, _ = fmt.Fprintf(w, "\n\n")
}

func outputSchedule(w io.Writer, rows [][]string, wait, turnaround, throughput float64) {
	_, _ = fmt.Fprintln(w, "Schedule table")
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"ID", "Priority", "Burst", "Arrival", "Wait", "Turnaround", "Exit"})
	table.AppendBulk(rows)
	table.SetFooter([]string{"", "", "", "",
		fmt.Sprintf("Average\n%.2f", wait),
		fmt.Sprintf("Average\n%.2f", turnaround),
		fmt.Sprintf("Throughput\n%.2f/t", throughput)})
	table.Render()
}

//endregion

//region Loading processes.

var ErrInvalidArgs = errors.New("invalid args")

func loadProcesses(r io.Reader) ([]Process, error) {
	rows, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: reading CSV", err)
	}

	processes := make([]Process, len(rows))
	for i := range rows {
		processes[i].ProcessID = mustStrToInt(rows[i][0])
		processes[i].BurstDuration = mustStrToInt(rows[i][1])
		processes[i].ArrivalTime = mustStrToInt(rows[i][2])
		if len(rows[i]) == 4 {
			processes[i].Priority = mustStrToInt(rows[i][3])
		}
	}

	return processes, nil
}

func mustStrToInt(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return i
}

//endregion
