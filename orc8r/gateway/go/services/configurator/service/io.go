/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

// package service implements the core of configurator
package service

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"magma/gateway/config"
	"magma/gateway/mconfig"
)

// SaveConfig saves new gateway.configs and returns old configuration if any
func SaveConfigs(cfgJson []byte, readOldCfg bool) (oldCfgJson []byte, err error) {
	if len(cfgJson) == 0 {
		return oldCfgJson, fmt.Errorf("empty gateway mconfigs")
	}
	mconfigPath := mconfig.ConfigFilePath()
	if readOldCfg {
		var oerr error
		if oldCfgJson, oerr = ioutil.ReadFile(mconfigPath); oerr != nil {
			oldCfgJson = nil
		}
	}
	err = safeSwap(mconfig.ConfigFilePath(), cfgJson)
	if err == nil {
		log.Printf("successfully updated mconfig in %s", mconfigPath)
	}
	return oldCfgJson, err
}

// updateStaticConfigs saves new gateway.configs into static mconfig location
func updateStaticConfigs(cfgJson []byte) error {
	intervalMin := config.GetMagmadConfigs().StaticMconfigUpdateIntervalMin
	if intervalMin <= 0 {
		return nil
	}
	intervalDuration := time.Duration(intervalMin) * time.Minute
	mconfigPath := mconfig.DefaultConfigFilePath()
	fi, err := os.Stat(mconfigPath)
	if (err == nil && fi.ModTime().Add(intervalDuration).Before(time.Now())) || os.IsNotExist(err) {
		return safeSwap(mconfigPath, cfgJson)
	}
	return nil
}

func safeSwap(mconfigPath string, cfgJson []byte) error {
	newMconfigPath := mconfigPath + ".new"
	oldMconfigPath := mconfigPath + ".old"
	err := ioutil.WriteFile(newMconfigPath, cfgJson, 0644)
	if err != nil {
		return fmt.Errorf("failed to save mconfigs into %s: %v", newMconfigPath, err)
	}
	oerr := os.Rename(mconfigPath, oldMconfigPath) // best effort, needed just for rollback on error
	err = os.Rename(newMconfigPath, mconfigPath)
	if err != nil {
		err = fmt.Errorf("failed to move mconfigs from %s to %s: %v", newMconfigPath, mconfigPath, err)
		if oerr == nil { // roll back if previous rename succeeded
			os.Rename(oldMconfigPath, mconfigPath)
		}
	}
	return err
}
