// +build soong

/*
 * Copyright 2019 Arm Limited.
 * SPDX-License-Identifier: Apache-2.0
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package core

import (
	"android/soong/android"

	"github.com/google/blueprint"
	"github.com/google/blueprint/proptools"
)

type nameProps struct {
	Name *string
}

type kernelModuleBackendProps struct {
	Srcs    []string
	Args    kbuildArgs
	Default bool
}

type kernelModuleBackend struct {
	android.ModuleBase
	Properties kernelModuleBackendProps
}

func kernelModuleBackendFactory() android.Module {
	m := &kernelModuleBackend{}

	m.AddProperties(&m.Properties)
	android.InitAndroidModule(m)

	return m
}

func (m *kernelModule) soongBuildActions(mctx android.TopDownMutatorContext) {

	nameProps := nameProps{proptools.StringPtr(m.Name())}

	props := kernelModuleBackendProps{
		Args:    m.generateKbuildArgs(mctx),
		Srcs:    m.Properties.getSources(mctx),
		Default: isBuiltByDefault(m),
	}

	mctx.CreateModule(android.ModuleFactoryAdaptor(kernelModuleBackendFactory), &nameProps, &props)
}

var soongKbuildRule = apctx.StaticRule("bobKbuild",
	blueprint.RuleParams{
		Command: "python $kmod_build -o $out --depfile $depfile " +
			"--common-root " + srcdir + " " +
			"--module-dir $output_module_dir $extra_includes " +
			"--sources $in $kbuild_extra_symbols " +
			"--kernel $kernel_dir --cross-compile '$kernel_cross_compile' " +
			"$cc_flag $hostcc_flag $clang_triple_flag " +
			"$kbuild_options --extra-cflags='$extra_cflags' $make_args",
		Depfile:     "$out.d",
		Deps:        blueprint.DepsGCC,
		Pool:        blueprint.Console,
		Description: "$out",
	}, "kmod_build", "depfile", "extra_includes", "extra_cflags", "kbuild_extra_symbols", "kernel_dir", "kernel_cross_compile",
	"kbuild_options", "make_args", "output_module_dir", "cc_flag", "hostcc_flag", "clang_triple_flag")

func (m *kernelModuleBackend) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	builtModule := android.PathForModuleOut(ctx, m.Name()+".ko")
	symvers := android.PathForModuleOut(ctx, "Module.symvers")

	ctx.Build(apctx,
		android.BuildParams{
			Rule:        soongKbuildRule,
			Description: "kbuild " + ctx.ModuleName(),
			Inputs:      android.PathsForSource(ctx, m.Properties.Srcs),
			//Implicits: utils.NewStringSlice(m.extraSymbolsFiles(ctx), []string{args["copy_with_deps"]}),
			Outputs: []android.WritablePath{builtModule},
			Args:    m.Properties.Args.toDict(),
			Default: m.Properties.Default,
		})

	ctx.InstallFile(android.PathForModuleInstall(ctx, "system", "vendor", "lib"),
		m.Name()+".ko", builtModule)

	// Add a dependency between Module.symvers and the kernel module. This
	// should really be added to Outputs or ImplicitOutputs above, but
	// Ninja doesn't support dependency files with multiple outputs yet.
	ctx.Build(apctx,
		android.BuildParams{
			Rule:    blueprint.Phony,
			Inputs:  []android.Path{builtModule},
			Outputs: []android.WritablePath{symvers},
		})
}