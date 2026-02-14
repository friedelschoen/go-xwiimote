package irpointer

func NewIRPointer() *IRPointer {
	return &IRPointer{
		Height: 384.0 / 512.0,

		MaxSbSlope: 0.7, // : tan(35 degrees)

		MinSbWidth: 0.1,

		SbWidth: 19.5, // cm

		SbDotWidth: 2.25, // cm

		SbDotHeight: 1.0, // cm

		SbOffScreenX: 0.0,
		SbOffScreenY: 0.0,

		// SbOffScreenX :  0.8f,
		// SbOffScreenY :  (0.8f * Height),
		// disable, may be doing more harm than good due to sensor pickup glitches

		SbSingleNoGuessDistance: (100.0 * 100.0),

		WiimoteFOVCoefficient: 0.39, // meter
	}
}

func NewErrorFilter() *ErrorFilter {
	return &ErrorFilter{
		ErrorMaxCount: 8,
	}
}

func NewGlitchFilter() *GlitchFilter {
	return &GlitchFilter{
		GlitchMaxCount: 5,
		GlitchDistance: (150.0 * 150.0),
	}
}

func NewMarcanSmoothing() *MarcanSmoothingFilter {
	return &MarcanSmoothingFilter{
		SmootherRadius:   8.0,
		SmootherSpeed:    0.25,
		SmootherDeadzone: 2.5, // pixels
	}
}

func NewSafeTopSensorNormalize() *TranslateFilter {
	return &TranslateFilter{
		Source:      FRect{FVec2{-340, -92}, FVec2{340, 290}},
		Destination: FRect{FVec2{-1, -1}, FVec2{1, 1}},
		Clamp:       true,
	}
}

func NewSafeBottomSensorNormalize() *TranslateFilter {
	return &TranslateFilter{
		Source:      FRect{FVec2{-340, -290}, FVec2{340, 92}},
		Destination: FRect{FVec2{-1, -1}, FVec2{1, 1}},
		Clamp:       true,
	}
}

func NewWideTopSensorNormalize() *TranslateFilter {
	return &TranslateFilter{
		Source:      FRect{FVec2{-430, -194}, FVec2{430, 290}},
		Destination: FRect{FVec2{-1, -1}, FVec2{1, 1}},
		Clamp:       true,
	}
}

func NewWideBottomSensorNormalize() *TranslateFilter {
	return &TranslateFilter{
		Source:      FRect{FVec2{-430, -290}, FVec2{430, 194}},
		Destination: FRect{FVec2{-1, -1}, FVec2{1, 1}},
		Clamp:       true,
	}
}
