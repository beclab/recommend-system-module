package common

import (
	"encoding/json"
	"log"
	"os"
)

func GetTaskFrequency() string {
	defaultFeq := "2"
	frequency := os.Getenv("TASK_FREQUENCY")
	if frequency == "" {
		return defaultFeq
	}
	return frequency
}

func GetAlgorithmVersion() string {
	defaultVersion := "latest"
	version := os.Getenv("ALGORITHM_VERSION")
	if version == "" {
		return defaultVersion
	}
	return version
}

func GetNameSpace() string {
	return os.Getenv("NAME_SPACE")
}

func GetRedisAddr() string {
	return os.Getenv("REDIS_ADDR")
}

func GetRedisPassword() string {
	return os.Getenv("REDIS_PASSOWRD")
}

func GetAppDataPath() string {
	return os.Getenv("APP_DATA_PATH")
}

func GetApplicationDataPath() string {
	return os.Getenv("APPLICATION_DATA_PATH")
}

func GetTermiusUserName() string {
	return os.Getenv("TERMIUS_USER_NAME")
}

func GetArgoUrl() string {
	defaultUrl := "http://localhost:2746/api"
	argoUrl := os.Getenv("ARGO_URL")
	if argoUrl == "" {
		return defaultUrl
	}
	return argoUrl
}

type ArgoTemplatesStepData struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

type ArgoTemplatesContainerEnvMapKeyData struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type ArgoTemplatesContainerEnvValueFromData struct {
	ConfigMapKeyRef ArgoTemplatesContainerEnvMapKeyData `json:"configMapKeyRef"`
}

type ArgoTemplatesContainerEnvData struct {
	Name      string                                  `json:"name"`
	Value     string                                  `json:"value,omitempty"`
	ValueFrom *ArgoTemplatesContainerEnvValueFromData `json:"valueFrom,omitempty"`
}

type ArgoTemplatesContainerVolData struct {
	MountPath string `json:"mountPath"`
	Name      string `json:"name"`
}

type ArgoTemplatesContainerData struct {
	Image           string                          `json:"image"`
	ImagePullPolicy string                          `json:"imagePullPolicy"`
	Env             []ArgoTemplatesContainerEnvData `json:"env,omitempty"`
	VolumeMounts    []ArgoTemplatesContainerVolData `json:"volumeMounts,omitempty"`
}

type ArgoTemplatesData struct {
	Name      string                      `json:"name"`
	Steps     [][]ArgoTemplatesStepData   `json:"steps,omitempty"`
	Container *ArgoTemplatesContainerData `json:"container,omitempty"`
}
type ArgoVolumePathData struct {
	Type string `json:"type"`
	Path string `json:"path"`
}
type ArgoSpecVolumeData struct {
	Name     string             `json:"name"`
	HostPath ArgoVolumePathData `json:"hostPath"`
}

type ArgoSpecData struct {
	Schedule                   string `json:"schedule"`
	StartingDeadlineSeconds    int    `json:"startingDeadlineSeconds"`
	ConcurrencyPolicy          string `json:"concurrencyPolicy"`
	SuccessfulJobsHistoryLimit int    `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     int    `json:"failedJobsHistoryLimit"`
	Suspend                    bool   `json:"suspend"`

	TTLStrategy ArgoWorkflowTTLStrategy `json:"ttlStrategy"`

	WorkFlowSpec ArgoWorkflowSpecData `json:"workflowSpec"`
}

type ArgoWorkflowTTLStrategy struct {
	SuccessTTL    int `json:"secondsAfterSuccess"`
	CompletionTTL int `json:"secondsAfterCompletion"`
	FailureTTL    int `json:"secondsAfterFailure"`
}

type ArgoWorkflowSpecData struct {
	Entrypoint string               `json:"entrypoint"`
	Volumes    []ArgoSpecVolumeData `json:"volumes,omitempty"`
	Templates  []ArgoTemplatesData  `json:"templates"`
}

type ArgoMetaData struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
type ArgoWorkflowData struct {
	ApiVersion string       `json:"apiVersion"`
	Kind       string       `json:"kind"`
	Metadata   ArgoMetaData `json:"metadata"`
	Spec       ArgoSpecData `json:"spec"`
}

type ArgoData struct {
	Namespace    string           `json:"namespace"`
	ServerDryRun bool             `json:"serverDryRun"`
	Workflow     ArgoWorkflowData `json:"cronWorkflow"`
}

func GenerateArgoSyncPostData(nameSpace string) []byte {
	imagePullPolicy := "Always"

	var postObj ArgoData
	var workFlow ArgoWorkflowData
	workFlow.ApiVersion = "argoproj.io/v1alpha1"
	workFlow.Kind = "CronWorkflow"
	var metadata ArgoMetaData
	metadata.Namespace = nameSpace
	metadata.Name = "recommend-task-sync"
	workFlow.Metadata = metadata

	var workflowSpec ArgoWorkflowSpecData
	workflowSpec.Entrypoint = "syncFlow"
	var volum1 ArgoSpecVolumeData
	var volum2 ArgoSpecVolumeData
	var volum1Path ArgoVolumePathData
	var volum2Path ArgoVolumePathData
	volum1.Name = "nfs"
	volum1Path.Path = GetAppDataPath() + "/rss/data"
	volum1Path.Type = "DirectoryOrCreate"
	volum1.HostPath = volum1Path
	volum2.Name = "juicefs"
	volum2Path.Path = GetApplicationDataPath() + "/rss/data"
	volum2Path.Type = "DirectoryOrCreate"
	volum2.HostPath = volum2Path

	workflowSpec.Volumes = []ArgoSpecVolumeData{volum1, volum2}

	var templatesVol1 ArgoTemplatesContainerVolData
	templatesVol1.Name = "nfs"
	templatesVol1.MountPath = "/nfs"
	var templatesVol2 ArgoTemplatesContainerVolData
	templatesVol2.Name = "juicefs"
	templatesVol2.MountPath = "/juicefs"

	var temp1 ArgoTemplatesData
	var step1 ArgoTemplatesStepData
	step1.Name = "sync"
	step1.Template = "sync-template"
	temp1.Name = "syncFlow"
	temp1.Steps = [][]ArgoTemplatesStepData{{step1}}

	var termiusUserNameEnv ArgoTemplatesContainerEnvData
	termiusUserNameEnv.Name = "TERMIUS_USER_NAME"
	termiusUserNameEnv.Value = GetTermiusUserName()

	version := ":" + GetAlgorithmVersion()
	var temp2 ArgoTemplatesData
	var temp2Container ArgoTemplatesContainerData
	temp2Container.Image = "beclab/recommend-sync" + version
	temp2Container.ImagePullPolicy = imagePullPolicy
	temp2Container.VolumeMounts = []ArgoTemplatesContainerVolData{templatesVol1, templatesVol2}
	temp2Container.Env = []ArgoTemplatesContainerEnvData{getMongoUrlEnv(), getMongoDbEnv(), getRedisAddEnv(), getRedisPasswordEnv(), getNfsDirectoryEnv(), getJuicefsDirectoryEnv(), getKnowledgeBaseApiUrlEnv(), getConfigFileEnv(), termiusUserNameEnv}
	temp2.Name = "sync-template"
	temp2.Container = &temp2Container

	workflowSpec.Templates = []ArgoTemplatesData{temp1, temp2}

	var spec ArgoSpecData

	spec.Schedule = "7/10 * * * *"
	spec.StartingDeadlineSeconds = 0
	spec.ConcurrencyPolicy = "Replace"
	spec.SuccessfulJobsHistoryLimit = 1
	spec.FailedJobsHistoryLimit = 1
	spec.TTLStrategy = getArgoFlowTTL()
	spec.Suspend = false
	spec.WorkFlowSpec = workflowSpec

	workFlow.Spec = spec

	postObj.Namespace = nameSpace
	postObj.ServerDryRun = false
	postObj.Workflow = workFlow

	body, err := json.Marshal(postObj)
	if err != nil {
		log.Print("Marshal data  fail", err)
	}
	return body
}

func GenerateArgoCrawlercPostData(nameSpace string) []byte {
	imagePullPolicy := "Always"

	var postObj ArgoData
	var workFlow ArgoWorkflowData
	workFlow.ApiVersion = "argoproj.io/v1alpha1"
	workFlow.Kind = "CronWorkflow"
	var metadata ArgoMetaData
	metadata.Namespace = nameSpace
	metadata.Name = "recommend-task-crawler"
	workFlow.Metadata = metadata

	var workflowSpec ArgoWorkflowSpecData
	workflowSpec.Entrypoint = "crawlerFlow"

	var temp1 ArgoTemplatesData
	var step1 ArgoTemplatesStepData
	step1.Name = "crawler"
	step1.Template = "crawler-template"
	temp1.Name = "crawlerFlow"
	temp1.Steps = [][]ArgoTemplatesStepData{{step1}}

	var termiusUserNameEnv ArgoTemplatesContainerEnvData
	termiusUserNameEnv.Name = "TERMIUS_USER_NAME"
	termiusUserNameEnv.Value = GetTermiusUserName()

	version := ":" + GetAlgorithmVersion()
	var temp2 ArgoTemplatesData
	var temp2Container ArgoTemplatesContainerData
	temp2Container.Image = "beclab/recommend-crawler" + version
	temp2Container.ImagePullPolicy = imagePullPolicy
	temp2Container.Env = []ArgoTemplatesContainerEnvData{getKnowledgeBaseApiUrlEnv(), termiusUserNameEnv}
	temp2.Name = "crawler-template"
	temp2.Container = &temp2Container

	workflowSpec.Templates = []ArgoTemplatesData{temp1, temp2}

	var spec ArgoSpecData
	spec.Schedule = "*/4 * * * *"
	spec.StartingDeadlineSeconds = 0
	spec.ConcurrencyPolicy = "Forbid"
	spec.SuccessfulJobsHistoryLimit = 1
	spec.FailedJobsHistoryLimit = 1
	spec.TTLStrategy = getArgoFlowTTL()
	spec.Suspend = false
	spec.WorkFlowSpec = workflowSpec

	workFlow.Spec = spec

	postObj.Namespace = nameSpace
	postObj.ServerDryRun = false
	postObj.Workflow = workFlow

	body, err := json.Marshal(postObj)
	if err != nil {
		log.Print("Marshal data  fail", err)
	}
	return body
}

func getMongoUrlEnv() ArgoTemplatesContainerEnvData {
	var mongoUrlEnv ArgoTemplatesContainerEnvData
	var mongoUrlEnvValueFrom ArgoTemplatesContainerEnvValueFromData
	var mongoUrlEnvMapKey ArgoTemplatesContainerEnvMapKeyData
	mongoUrlEnvMapKey.Name = "rss-secrets-auth"
	mongoUrlEnvMapKey.Key = "mongo_url"
	mongoUrlEnvValueFrom.ConfigMapKeyRef = mongoUrlEnvMapKey
	mongoUrlEnv.Name = "TERMINUS_RECOMMEND_MONGODB_URI"
	mongoUrlEnv.ValueFrom = &mongoUrlEnvValueFrom
	return mongoUrlEnv
}

func getMongoDbEnv() ArgoTemplatesContainerEnvData {
	var mongoDbEnv ArgoTemplatesContainerEnvData
	var mongoDbEnvValueFrom ArgoTemplatesContainerEnvValueFromData
	var mongoDbEnvMapKey ArgoTemplatesContainerEnvMapKeyData
	mongoDbEnvMapKey.Name = "rss-secrets-auth"
	mongoDbEnvMapKey.Key = "mongo_db"
	mongoDbEnvValueFrom.ConfigMapKeyRef = mongoDbEnvMapKey
	mongoDbEnv.Name = "TERMINUS_RECOMMEND_MONGODB_NAME"
	mongoDbEnv.ValueFrom = &mongoDbEnvValueFrom
	return mongoDbEnv
}

func getRedisAddEnv() ArgoTemplatesContainerEnvData {
	var redisAddEnv ArgoTemplatesContainerEnvData
	var redisAddEnvValueFrom ArgoTemplatesContainerEnvValueFromData
	var redisAddEnvMapKey ArgoTemplatesContainerEnvMapKeyData
	redisAddEnvMapKey.Name = "rss-secrets-auth"
	redisAddEnvMapKey.Key = "redis_addr"
	redisAddEnvValueFrom.ConfigMapKeyRef = redisAddEnvMapKey
	redisAddEnv.Name = "TERMINUS_RECOMMEND_REDIS_ADDR"
	//redisAddEnv.Value = GetRedisAddr()
	redisAddEnv.ValueFrom = &redisAddEnvValueFrom
	return redisAddEnv
}

func getRedisPasswordEnv() ArgoTemplatesContainerEnvData {
	var redisPasswordEnv ArgoTemplatesContainerEnvData
	var redisPasswordEnvValueFrom ArgoTemplatesContainerEnvValueFromData
	var redisPasswordEnvMapKey ArgoTemplatesContainerEnvMapKeyData
	redisPasswordEnvMapKey.Name = "rss-secrets-auth"
	redisPasswordEnvMapKey.Key = "redis_password"
	redisPasswordEnvValueFrom.ConfigMapKeyRef = redisPasswordEnvMapKey
	redisPasswordEnv.Name = "TERMINUS_RECOMMEND_REDIS_PASSOWRD"
	redisPasswordEnv.ValueFrom = &redisPasswordEnvValueFrom
	return redisPasswordEnv
}

func getNfsDirectoryEnv() ArgoTemplatesContainerEnvData {
	var nfsDirectoryEnv ArgoTemplatesContainerEnvData
	nfsDirectoryEnv.Name = "NFS_ROOT_DIRECTORY"
	nfsDirectoryEnv.Value = "/nfs"
	return nfsDirectoryEnv
}

func getJuicefsDirectoryEnv() ArgoTemplatesContainerEnvData {
	var juicefsDirectoryEnv ArgoTemplatesContainerEnvData
	juicefsDirectoryEnv.Name = "JUICEFS_ROOT_DIRECTORY"
	juicefsDirectoryEnv.Value = "/juicefs"
	return juicefsDirectoryEnv
}

func getKnowledgeBaseApiUrlEnv() ArgoTemplatesContainerEnvData {
	var knowledgeUrlEnv ArgoTemplatesContainerEnvData
	knowledgeUrlEnv.Name = "KNOWLEDGE_BASE_API_URL"
	knowledgeUrlEnv.Value = "http://knowledge-base-api.user-system-" + GetTermiusUserName() + ":" + os.Getenv("KNOWLEDGE_BASE_API_PORT") //os.Getenv("KNOWLEDGE_BASE_API_URL")
	return knowledgeUrlEnv
}

func getConfigFileEnv() ArgoTemplatesContainerEnvData {
	var configFileEnv ArgoTemplatesContainerEnvData
	configFileEnv.Name = "ALGORITHM_FILE_CONFIG_PATH"
	configFileEnv.Value = "/usr/config/"
	return configFileEnv
}

func getArgoFlowTTL() ArgoWorkflowTTLStrategy {
	var ttlStrategy ArgoWorkflowTTLStrategy
	ttlStrategy.SuccessTTL = 5
	ttlStrategy.CompletionTTL = 10
	ttlStrategy.FailureTTL = 0
	return ttlStrategy
}
