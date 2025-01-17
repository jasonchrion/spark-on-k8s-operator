package config

import (
	"fmt"
	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	"os"
	"strconv"
	"time"
)

const mntDir = "/mnt/"

type getCm struct {
	configMap *v1.ConfigMap
}

type getSecret struct {
	secret *v1.Secret
}

type dynamicCopy interface {
	copyResourceToFile(appNamespace string, appName string, resourceName string) (string, []string, error)
}

// Here, copyResourceToFile is a receiver function implementing dynamicCopy interface on getCm
// It copies configmap content to files and returns the location of the configmap files including the filename
func (gc getCm) copyResourceToFile(appNamespace string, appName string, cmName string) (string, []string, error) {
	var returnedpath string
	var filesInConfigMap []string
	for key := range gc.configMap.Data {
		pathVal := mntDir + appNamespace + "/" + appName + "/" + cmName
		err := os.MkdirAll(pathVal, 0770)
		if err != nil {
			glog.Errorf("%v", err)
			return "", nil, err
		}

		pathValKey := pathVal + "/" + key
		f, err := os.Create(pathValKey)
		if err != nil {
			glog.Errorf("%v", err)
			return "", nil, err
		} else {
			glog.V(2).Infof("%s created successfully", pathValKey)
		}

		defer f.Close()
		f.WriteString(gc.configMap.Data[key])
		filesInConfigMap = append(filesInConfigMap, key)
		returnedpath = pathVal
	}
	return returnedpath, filesInConfigMap, nil
}

// Here, copyResourceToFile is a receiver function implementing dynamicCopy interface on getSecret
// It copies secret content to file and returns the location of the file including the filename
func (gs getSecret) copyResourceToFile(appNamespace string, appName string, secretName string) (string, []string, error) {
	var returnedpath string
	var filesInSecret []string
	for key := range gs.secret.Data {
		pathVal := mntDir + appNamespace + "/" + appName + "/" + secretName
		err := os.MkdirAll(pathVal, 0770)
		if err != nil {
			glog.Errorf("%v", err)
			return "", nil, err
		}

		pathValKey := pathVal + "/" + key
		f, err := os.Create(pathValKey)
		if err != nil {
			glog.Errorf("%v", err)
			return "", nil, err
		} else {
			glog.V(2).Infof("%s created successfully", pathValKey)
		}

		defer f.Close()
		f.WriteString(string(gs.secret.Data[key]))
		filesInSecret = append(filesInSecret, key)
		returnedpath = pathVal
	}
	return returnedpath, filesInSecret, nil
}

// If a variable has an interface type, then we can call methods that are in the named interface
// So, instance of getCm struct can be used as argument to copyToFile function
func copyToFile(d dynamicCopy, appNamespace string, appName string, resourceName string) (string, error) {
	path, key, err := d.copyResourceToFile(appNamespace, appName, resourceName)
	if err != nil {
		return "", err
	}
	_ = key
	return path, nil
}

// RemoveDirectory removes the directory where configmaps and secrets are dynamically copied
func RemoveDirectory(appNamespace string, appName string) {
	var dirToRemove = mntDir + appNamespace + "/" + appName
	debugMode, present := os.LookupEnv("DEBUG_MODE")
	if present && debugMode == "true" {
		glog.V(2).Infof("Debug mode enabled, DO NOT DELETE %s", dirToRemove)
	} else {
		err := os.RemoveAll(dirToRemove)
		if err != nil {
			glog.Errorf("%v", err)
		} else {
			glog.V(2).Infof("Deleting %s if it exists", dirToRemove)
		}
	}
}

func WriteSparkSubmitShell(appNamespace string, appName string, commands []string) (*os.File, error) {
	var dir = mntDir + appNamespace + "/" + appName
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0770)
		if err != nil {
			glog.Errorf("mkdir %s error: %v", dir, err)
			return nil, err
		}
	}
	shellFile := dir + "/" + appName + strconv.FormatInt(time.Now().Unix(), 10) + ".sh"
	var file *os.File
	if _, err := os.Stat(shellFile); os.IsNotExist(err) {
		file, err = os.Create(shellFile)
		if err != nil {
			glog.Errorf("create file %s error: %v", shellFile, err)
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("shell file for SparkApplication %s/%s already exists", appNamespace, appName)
	}
	defer file.Close()
	os.Chmod(file.Name(), 0770)
	for _, command := range commands {
		file.WriteString(command + "\n")
	}
	return file, nil
}
