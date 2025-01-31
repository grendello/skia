/*
 * Copyright 2022 Google LLC
 *
 * Use of this source code is governed by a BSD-style license that can be
 * found in the LICENSE file.
 */

#include "src/shaders/SkEmptyShader.h"

#include "include/core/SkFlattenable.h"
#include "include/core/SkRefCnt.h"
#include "include/core/SkShader.h"

class SkReadBuffer;

#if defined(SK_ENABLE_SKVM)
skvm::Color SkEmptyShader::program(skvm::Builder*,
                                   skvm::Coord,
                                   skvm::Coord,
                                   skvm::Color,
                                   const SkShaders::MatrixRec&,
                                   const SkColorInfo&,
                                   skvm::Uniforms*,
                                   SkArenaAlloc*) const {
    return {};  // signal failure
}
#endif

sk_sp<SkFlattenable> SkEmptyShader::CreateProc(SkReadBuffer&) {
    return SkShaders::Empty();
}

sk_sp<SkShader> SkShaders::Empty() { return sk_make_sp<SkEmptyShader>(); }

void SkRegisterEmptyShaderFlattenable() {
    SK_REGISTER_FLATTENABLE(SkEmptyShader);
}
