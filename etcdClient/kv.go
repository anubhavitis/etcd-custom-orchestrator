package etcdClient

import (
	"context"
	"encoding/json"
	"fmt"

	etcd "go.etcd.io/etcd/client/v3"
)

var Client *etcd.Client

func init() {
	config := etcd.Config{
		Endpoints: []string{"http://localhost:2379"}, // Etcd server endpoints
	}
	client, err := etcd.New(config)
	if err != nil {
		fmt.Println("Error connecting to etcd:", err)
		return
	}

	Client = client
}

func GetClient() *etcd.Client {
	return Client
}

func TriggerUpdate(jobs []Job) {
	/*
		Update the list of jobs that needs to be refreshed
	*/
	updateJobNames := make([]string, len(jobs))
	for i, job := range jobs {
		updateJobNames[i] = job.Name
	}

	jobsBytes, _ := json.Marshal(updateJobNames)

	Client.Put(context.TODO(), JobsUpdatedListKey, string(jobsBytes))
}

func GetEtcdJobMap() map[string]Job {
	jobs := GetEtcdJobs()
	resp := make(map[string]Job)

	for i := range jobs {
		resp[jobs[i].Name] = jobs[i]
	}
	return resp
}

func GetEtcdJobs() (resp []Job) {
	/*
		Get list of all the jobs
	*/
	getResp, err := Client.Get(context.Background(), JobConfigKey)
	if err != nil {
		fmt.Println("Error getting etcd jobs, err: ", err)
	}

	Kvs := getResp.Kvs
	fmt.Println("Key values found received: ", len(Kvs))
	for i := range Kvs {
		key := string(Kvs[i].Key)
		value := string(Kvs[i].Value)

		fmt.Println(" key: ", key, " and value: ", value)
	}

	if len(Kvs) == 0 {
		return []Job{}
	}

	value := Kvs[0].Value
	json.Unmarshal([]byte(value), &resp)
	return
}

func UpdateJobsConfig(job Job) (prevJ Job, err error) {
	/*
		Append the new Job in the list of jobs
	*/
	key := JobConfigKey
	currentJobs := GetEtcdJobs()
	newJobs := append(currentJobs, job)
	jobsByte, e := json.Marshal(newJobs)
	if e != nil {
		fmt.Println("error while marshalling for key: ", key, "due to error:", e)
		return prevJ, e
	}

	resp, err := Client.Put(context.Background(), key, string(jobsByte))
	if err != nil {
		fmt.Println("PUT operation failed for key: ", key, "due to error:", e)
		return prevJ, e
	}

	TriggerUpdate([]Job{job})
	fmt.Println("Previous Value:", resp)
	fmt.Println("Privious Kv:", resp.PrevKv)
	// previousValue := resp.PrevKv.Value
	// err = json.Unmarshal(previousValue, &prevJ)
	return
}

func GetKeyHistory(key string) ([]Job, error) {
	resp, err := Client.Get(context.Background(), key, etcd.WithPrefix(), etcd.WithSort(etcd.SortByCreateRevision, etcd.SortAscend))
	if err != nil {
		fmt.Println("Error getting key history:", err)
		return nil, err
	}

	// Extract and parse historical values into Job struct
	history := make([]Job, 0)

	for _, kv := range resp.Kvs {
		var job Job
		if err := json.Unmarshal(kv.Value, &job); err != nil {
			fmt.Printf("Error unmarshalling key %s with revision %d: %v\n", key, kv.CreateRevision, err)
			continue
		}
		history = append(history, job)
	}

	return history, nil

}
