package com.hailam.malgoplay

import androidx.lifecycle.ViewModel
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import mobile_fsg_main.Mobile_fsg_main

class MainViewModel : ViewModel() {

    companion object {
        const val DEFAULT_MAX_FREQUENCY = 1000.0
        const val DEFAULT_MIN_FREQUENCY = 600.0
        const val DEFAULT_AMPLITUDE = 2.0
        const val DEFAULT_SWEEP_RATE = 1.0
        const val DEFAULT_CHANNELS = 1L
    }

    private var isPlaying = false

    fun startPlaying(
        minFrequency: Double,
        maxFrequency: Double,
        amplitude: Double,
        sweepRate: Double,
        channels: Long,
        mode: String
    ) {
        CoroutineScope(Dispatchers.IO).launch {
            // Initialize the audio device
            initAudioDevice(minFrequency, maxFrequency, channels)
            Mobile_fsg_main.setSweepRate(sweepRate)
            Mobile_fsg_main.setAmplitude(amplitude)

            val selectedMode = when (mode) {
                "Linear" -> 0L
                "Triangle" -> 1L
                "Sine" -> 2L
                "Exponential" -> 3L
                "Logarithmic" -> 4L
                "Square" -> 5L
                "Sawtooth" -> 6L
                "Random" -> 7L
                else -> 2L // Default to Sine
            }

            Mobile_fsg_main.setSweepMode(selectedMode)
            Mobile_fsg_main.startAudio(0)

            withContext(Dispatchers.Main) {
                isPlaying = true
            }
        }
    }

    fun stopPlaying() {
        CoroutineScope(Dispatchers.IO).launch {
            Mobile_fsg_main.stopAudio()
            withContext(Dispatchers.Main) {
                isPlaying = false
            }
            Mobile_fsg_main.cleanupAudio()
        }
    }

    private fun initAudioDevice(minFrequency: Double, maxFrequency: Double, channels: Long) {
        val sampleRate = 44100L
        Mobile_fsg_main.initializeAudio(minFrequency, maxFrequency, sampleRate, channels)
    }
}
