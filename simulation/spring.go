package simulation

import (
	"math"
	"time"
)

var epsilon = math.Nextafter(1, 2) - 1

// FPS returns the delta time for a fixed frames-per-second rate.
func FPS(framesPerSecond int) float64 {
	return (time.Second / time.Duration(framesPerSecond)).Seconds()
}

// Spring caches coefficients for a damped harmonic oscillator.
type Spring struct {
	posPosCoef float64
	posVelCoef float64
	velPosCoef float64
	velVelCoef float64
}

// NewSpring computes the coefficients for a spring with the given time step,
// angular frequency, and damping ratio.
func NewSpring(deltaTime, angularFrequency, dampingRatio float64) (spring Spring) {
	angularFrequency = math.Max(0.0, angularFrequency)
	dampingRatio = math.Max(0.0, dampingRatio)

	if angularFrequency < epsilon {
		spring.posPosCoef = 1.0
		spring.velVelCoef = 1.0
		return spring
	}

	if dampingRatio > 1.0+epsilon {
		baseDecay := -angularFrequency * dampingRatio
		dampingOffset := angularFrequency * math.Sqrt(dampingRatio*dampingRatio-1.0)
		decayRoot1 := baseDecay - dampingOffset
		decayRoot2 := baseDecay + dampingOffset

		exp1 := math.Exp(decayRoot1 * deltaTime)
		exp2 := math.Exp(decayRoot2 * deltaTime)

		inverseDoubleOffset := 1.0 / (2.0 * dampingOffset)
		exp1OverDoubleOffset := exp1 * inverseDoubleOffset
		exp2OverDoubleOffset := exp2 * inverseDoubleOffset
		decayRoot1Exp1OverDoubleOffset := decayRoot1 * exp1OverDoubleOffset
		decayRoot2Exp2OverDoubleOffset := decayRoot2 * exp2OverDoubleOffset

		spring.posPosCoef = exp1OverDoubleOffset*decayRoot2 - decayRoot2Exp2OverDoubleOffset + exp2
		spring.posVelCoef = -exp1OverDoubleOffset + exp2OverDoubleOffset
		spring.velPosCoef = (decayRoot1Exp1OverDoubleOffset - decayRoot2Exp2OverDoubleOffset + exp2) * decayRoot2
		spring.velVelCoef = -decayRoot1Exp1OverDoubleOffset + decayRoot2Exp2OverDoubleOffset
	} else if dampingRatio < 1.0-epsilon {
		omegaZeta := angularFrequency * dampingRatio
		alpha := angularFrequency * math.Sqrt(1.0-dampingRatio*dampingRatio)

		expTerm := math.Exp(-omegaZeta * deltaTime)
		cosTerm := math.Cos(alpha * deltaTime)
		sinTerm := math.Sin(alpha * deltaTime)
		invAlpha := 1.0 / alpha
		expSin := expTerm * sinTerm
		expCos := expTerm * cosTerm
		expOmegaZetaSinOverAlpha := expTerm * omegaZeta * sinTerm * invAlpha

		spring.posPosCoef = expCos + expOmegaZetaSinOverAlpha
		spring.posVelCoef = expSin * invAlpha
		spring.velPosCoef = -expSin*alpha - omegaZeta*expOmegaZetaSinOverAlpha
		spring.velVelCoef = expCos - expOmegaZetaSinOverAlpha
	} else {
		expTerm := math.Exp(-angularFrequency * deltaTime)
		timeExp := deltaTime * expTerm
		timeExpFreq := timeExp * angularFrequency

		spring.posPosCoef = timeExpFreq + expTerm
		spring.posVelCoef = timeExp
		spring.velPosCoef = -angularFrequency * timeExpFreq
		spring.velVelCoef = -timeExpFreq + expTerm
	}

	return spring
}

// Update advances a position and velocity toward the given equilibrium point.
func (spring Spring) Update(pos, vel, equilibriumPos float64) (newPos, newVel float64) {
	oldPos := pos - equilibriumPos
	oldVel := vel

	newPos = oldPos*spring.posPosCoef + oldVel*spring.posVelCoef + equilibriumPos
	newVel = oldPos*spring.velPosCoef + oldVel*spring.velVelCoef

	return newPos, newVel
}
