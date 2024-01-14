package strategy

type AllocStrategy func(jobKey string, myNodeId string, allNodeIds []string) (bool, error)
