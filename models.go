package main

import "time"

type Job struct {
	RequestID   int       `json:"requestId"`
	QueueTime   time.Time `json:"queueTime"`
	AssignTime  time.Time `json:"assignTime"`
	ReceiveTime time.Time `json:"receiveTime,omitempty"`
	FinishTime  time.Time `json:"finishTime,omitempty"`
	Result      string    `json:"result,omitempty"`
}

type Jobs struct {
	Count int   `json:"count"`
	Value []Job `json:"value"`
}

type Agent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Agents struct {
	Count  int     `json:"count"`
	Agents []Agent `json:"value"`
}
