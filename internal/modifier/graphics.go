/**
# Copyright (c) 2022, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package modifier

import (
	"fmt"
	"strings"

	"github.com/NVIDIA/nvidia-container-toolkit/internal/config"
	"github.com/NVIDIA/nvidia-container-toolkit/internal/config/image"
	"github.com/NVIDIA/nvidia-container-toolkit/internal/discover"
	"github.com/NVIDIA/nvidia-container-toolkit/internal/oci"
	"github.com/sirupsen/logrus"
)

// NewGraphicsModifier constructs a modifier that injects graphics-related modifications into an OCI runtime specification.
// The value of the NVIDIA_DRIVER_CAPABILITIES environment variable is checked to determine if this modification should be made.
func NewGraphicsModifier(logger *logrus.Logger, cfg *config.Config, ociSpec oci.Spec) (oci.SpecModifier, error) {
	rawSpec, err := ociSpec.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load OCI spec: %v", err)
	}

	image, err := image.NewCUDAImageFromSpec(rawSpec)
	if err != nil {
		return nil, err
	}

	if devices := image.DevicesFromEnvvars(visibleDevicesEnvvar); len(devices) == 0 {
		logger.Infof("No modification required; no devices requested")
		return nil, nil
	}

	var hasGraphics bool
	for _, c := range strings.Split(image["NVIDIA_DRIVER_CAPABILITIES"], ",") {
		if c == "graphics" || c == "all" {
			hasGraphics = true
			break
		}
	}

	if !hasGraphics {
		logger.Debugf("Capability %q not selected", "graphics")
		return nil, nil
	}

	d, err := discover.NewGraphicsDiscoverer(logger, cfg.NVIDIAContainerCLIConfig.Root)
	if err != nil {
		return nil, fmt.Errorf("failed to construct discoverer: %v", err)
	}

	return NewModifierFromDiscoverer(logger, d)
}
