// Copyright (C) 2025 ANSYS, Inc. and/or its affiliates.
// SPDX-License-Identifier: MIT
//
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package azure

import (
	"fmt"
	"strconv"

	"github.com/ansys/aali-sharedtypes/pkg/config"
	"github.com/ansys/aali-sharedtypes/pkg/logging"
)

func mustCfg(ctx *logging.ContextMap, key string) string {
	toolVal, ok := config.GlobalConfig.WORKFLOW_CONFIG_VARIABLES[key]

	if !ok {
		err := fmt.Sprintf("%s not found in configuration", key)
		logging.Log.Error(ctx, err)
		panic(err)
	}

	return toolVal
}

// GetSubworkflows returns a slice of subworkflow names and descriptions.
func GetSubworkflows() []struct {
	Name        string
	Description string
} {
	ctx := &logging.ContextMap{}
	subworkflows := []struct {
		Name        string
		Description string
	}{}

	toolAmount := mustCfg(ctx, "APP_TOOL_TOTAL_AMOUNT")
	toolAmountInt, err := strconv.Atoi(toolAmount)
	if err != nil {
		logging.Log.Error(ctx, fmt.Sprintf("Invalid tool amount: %s", toolAmount))
		panic(err)
	}

	for i := 1; i <= toolAmountInt; i++ {
		nameKey := fmt.Sprintf("APP_TOOL_%d_NAME", i)
		descKey := fmt.Sprintf("APP_TOOL_%d_DESCRIPTION", i)
		name, nameExists := mustCfg(ctx, nameKey), true
		desc, descExists := mustCfg(ctx, descKey), true
		if nameExists && descExists && name != "" && desc != "" {
			subworkflows = append(subworkflows, struct {
				Name        string
				Description string
			}{
				Name:        name,
				Description: desc,
			})
		}
	}
	return subworkflows
}
