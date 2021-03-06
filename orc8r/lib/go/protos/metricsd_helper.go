/*
Copyright (c) Facebook, Inc. and its affiliates.
All rights reserved.

This source code is licensed under the BSD-style license found in the
LICENSE file in the root directory of this source tree.
*/

// metricsd_helper.go adds some useful conversions between metric/label enum
// values and names

package protos

import (
	"strconv"

	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
)

// GetEnumNameIfPossible tries to convert a string enum value to the associated
// enum name. If toConvert doesn't match up with any mapping in protoEnumMapping
// then this function just returns the original toConvert string.
func GetEnumNameIfPossible(toConvert string,
	protoEnumMapping map[int32]string) string {
	enumVal, err := strconv.Atoi(toConvert)
	nameStr := ""
	if err == nil {
		nameStr = proto.EnumName(protoEnumMapping, int32(enumVal))
	} else {
		nameStr = toConvert
	}
	return nameStr
}

// GetDecodedLabel tries to convert the metric label name/value enums to their
// enum names for display.
func GetDecodedLabel(m *dto.Metric) []*dto.LabelPair {
	var newLabels []*dto.LabelPair
	for _, labelPair := range m.GetLabel() {
		labelName := GetEnumNameIfPossible(
			labelPair.GetName(),
			MetricLabelName_name)
		labelValue := labelPair.GetValue()
		newLabels = append(
			newLabels,
			&dto.LabelPair{Name: &labelName, Value: &labelValue})
	}
	return newLabels
}

// GetDecodedName gets the enum name for the metric family from the enum value
func GetDecodedName(m *dto.MetricFamily) string {
	return GetEnumNameIfPossible(m.GetName(), MetricName_name)
}
