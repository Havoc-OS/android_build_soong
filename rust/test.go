// Copyright 2019 The Android Open Source Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rust

import (
	"android/soong/android"
	"android/soong/tradefed"
)

type TestProperties struct {
	// Disables the creation of a test-specific directory when used with
	// relative_install_path. Useful if several tests need to be in the same
	// directory, but test_per_src doesn't work.
	No_named_install_directory *bool

	// the name of the test configuration (for example "AndroidTest.xml") that should be
	// installed with the module.
	Test_config *string `android:"path,arch_variant"`

	// the name of the test configuration template (for example "AndroidTestTemplate.xml") that
	// should be installed with the module.
	Test_config_template *string `android:"path,arch_variant"`

	// list of compatibility suites (for example "cts", "vts") that the module should be
	// installed into.
	Test_suites []string `android:"arch_variant"`

	// Flag to indicate whether or not to create test config automatically. If AndroidTest.xml
	// doesn't exist next to the Android.bp, this attribute doesn't need to be set to true
	// explicitly.
	Auto_gen_config *bool
}

// A test module is a binary module with extra --test compiler flag
// and different default installation directory.
// In golang, inheriance is written as a component.
type testDecorator struct {
	*binaryDecorator
	Properties TestProperties
	testConfig android.Path
}

func (test *testDecorator) nativeCoverage() bool {
	return true
}

func NewRustTest(hod android.HostOrDeviceSupported) (*Module, *testDecorator) {
	// Build both 32 and 64 targets for device tests.
	// Cannot build both for host tests yet if the test depends on
	// something like proc-macro2 that cannot be built for both.
	multilib := android.MultilibBoth
	if hod != android.DeviceSupported && hod != android.HostAndDeviceSupported {
		multilib = android.MultilibFirst
	}
	module := newModule(hod, multilib)

	test := &testDecorator{
		binaryDecorator: &binaryDecorator{
			baseCompiler: NewBaseCompiler("nativetest", "nativetest64", InstallInData),
		},
	}

	module.compiler = test
	module.AddProperties(&test.Properties)
	return module, test
}

func (test *testDecorator) compilerProps() []interface{} {
	return append(test.binaryDecorator.compilerProps(), &test.Properties)
}

func (test *testDecorator) install(ctx ModuleContext, file android.Path) {
	test.testConfig = tradefed.AutoGenRustTestConfig(ctx,
		test.Properties.Test_config,
		test.Properties.Test_config_template,
		test.Properties.Test_suites,
		nil,
		test.Properties.Auto_gen_config)

	// default relative install path is module name
	if !Bool(test.Properties.No_named_install_directory) {
		test.baseCompiler.relative = ctx.ModuleName()
	} else if String(test.baseCompiler.Properties.Relative_install_path) == "" {
		ctx.PropertyErrorf("no_named_install_directory", "Module install directory may only be disabled if relative_install_path is set")
	}

	test.binaryDecorator.install(ctx, file)
}

func (test *testDecorator) compilerFlags(ctx ModuleContext, flags Flags) Flags {
	flags = test.binaryDecorator.compilerFlags(ctx, flags)
	flags.RustFlags = append(flags.RustFlags, "--test")
	return flags
}

func (test *testDecorator) autoDep() autoDep {
	return rlibAutoDep
}

func init() {
	// Rust tests are binary files built with --test.
	android.RegisterModuleType("rust_test", RustTestFactory)
	android.RegisterModuleType("rust_test_host", RustTestHostFactory)
}

func RustTestFactory() android.Module {
	module, _ := NewRustTest(android.HostAndDeviceSupported)
	return module.Init()
}

func RustTestHostFactory() android.Module {
	module, _ := NewRustTest(android.HostSupported)
	return module.Init()
}