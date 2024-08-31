package com.hailam.malgoplay

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.view.View
import android.widget.*
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import mobile_fsg_main.Mobile_fsg_main

class MainActivity : AppCompatActivity() {

    companion object {
        private const val REQUEST_AUDIO_PERMISSION = 1
        private const val DEFAULT_MAX_FREQUENCY = 1000.0
        private const val DEFAULT_MIN_FREQUENCY = 600.0
        private const val DEFAULT_AMPLITUDE = 2.0
        private const val DEFAULT_SWEEP_RATE = 1.0

        private const val SWEEP_MODE_LINEAR = 0L
        private const val SWEEP_MODE_SINE = 1L
        private const val SWEEP_MODE_TRIANGLE = 2L
        private const val SWEEP_MODE_EXPONENTIAL = 3L
        private const val SWEEP_MODE_LOGARITHMIC = 4L
        private const val SWEEP_MODE_SQUARE = 5L
        private const val SWEEP_MODE_SAWTOOTH = 6L
        private const val SWEEP_MODE_RANDOM = 7L
    }

    private var isPlaying = false
    private lateinit var modeSpinner: Spinner
    private lateinit var maxFrequencyInput: EditText
    private lateinit var amplitudeInput: EditText
    private lateinit var minFrequencyInput: EditText
    private lateinit var sweepRateInput: EditText
    private lateinit var playButton: Button
    private lateinit var animationContainer: FrameLayout

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        if (ContextCompat.checkSelfPermission(this, Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.RECORD_AUDIO), REQUEST_AUDIO_PERMISSION)
        }

        // Get references to the UI elements
        modeSpinner = findViewById(R.id.mode_spinner)
        maxFrequencyInput = findViewById(R.id.frequency_input)
        amplitudeInput = findViewById(R.id.amplitude_input)
        minFrequencyInput = findViewById(R.id.min_frequency_input)
        sweepRateInput = findViewById(R.id.sweep_rate_input)
        playButton = findViewById(R.id.play_button)

        // Initialize Spinner with modes
        val modes = listOf("Sine", "Linear", "Triangle", "Exponential", "Logarithmic", "Square", "Sawtooth", "Random")
        val adapter = ArrayAdapter(this, android.R.layout.simple_spinner_item, modes)
        adapter.setDropDownViewResource(android.R.layout.simple_spinner_dropdown_item)
        modeSpinner.adapter = adapter

        // Set default values
        maxFrequencyInput.setText(DEFAULT_MAX_FREQUENCY.toString())
        amplitudeInput.setText(DEFAULT_AMPLITUDE.toString())
        minFrequencyInput.setText(DEFAULT_MIN_FREQUENCY.toString())
        sweepRateInput.setText(DEFAULT_SWEEP_RATE.toString())

        // Set an onClick listener for the Play/Stop button
        playButton.setOnClickListener {
            if (!isPlaying) {
                startPlaying()
            } else {
                stopPlaying()
            }
        }
    }

    private fun startPlaying() {
        val maxFrequency = maxFrequencyInput.text.toString().toDoubleOrNull() ?: DEFAULT_MAX_FREQUENCY
        val minFrequency = minFrequencyInput.text.toString().toDoubleOrNull() ?: DEFAULT_MIN_FREQUENCY
        val amplitude = amplitudeInput.text.toString().toDoubleOrNull() ?: DEFAULT_AMPLITUDE
        val sweepRate = sweepRateInput.text.toString().toDoubleOrNull() ?: DEFAULT_SWEEP_RATE

        // Initialize the audio device with the current user input values
        initAudioDevice(minFrequency, maxFrequency)

        // Set sweep rate and mode before starting the audio
        Mobile_fsg_main.setSweepRate(sweepRate)
        val selectedMode = when (modeSpinner.selectedItem.toString()) {
            "Linear" -> SWEEP_MODE_LINEAR
            "Triangle" -> SWEEP_MODE_TRIANGLE
            "Sine" -> SWEEP_MODE_SINE
            "Exponential" -> SWEEP_MODE_EXPONENTIAL
            "Logarithmic" -> SWEEP_MODE_LOGARITHMIC
            "Square" -> SWEEP_MODE_SQUARE
            "Sawtooth" -> SWEEP_MODE_SAWTOOTH
            "Random" -> SWEEP_MODE_RANDOM
            else -> SWEEP_MODE_SINE // Default to Sine
        }
        Mobile_fsg_main.setSweepMode(selectedMode)
        Mobile_fsg_main.startAudio(0) // 0 for indefinite playback

        // Toggle UI for playing state
        isPlaying = true
        playButton.text = "Stop"

        // Hide input fields and show animation
        /*
        maxFrequencyInput.visibility = View.GONE
        amplitudeInput.visibility = View.GONE
        minFrequencyInput.visibility = View.GONE
        sweepRateInput.visibility = View.GONE
        modeSpinner.visibility = View.GONE
         */

        // TODO: Start the animation or graphic in the animationContainer
    }

    private fun stopPlaying() {
        Mobile_fsg_main.stopAudio()

        // Toggle UI for stopped state
        playButton.text = "Play"
        isPlaying = false

        // Show input fields and hide animation
        /*
        maxFrequencyInput.visibility = View.VISIBLE
        amplitudeInput.visibility = View.VISIBLE
        minFrequencyInput.visibility = View.VISIBLE
        sweepRateInput.visibility = View.VISIBLE
        modeSpinner.visibility = View.VISIBLE
        */

        // Reset state in the generator
        Mobile_fsg_main.cleanupAudio()

        // TODO: Stop the animation or graphic in the animationContainer
    }


    private fun initAudioDevice(minFrequency: Double, maxFrequency: Double) {
        val sampleRate = 44100L
        val channels = 1L

        val result = Mobile_fsg_main.initializeAudio(minFrequency, maxFrequency, sampleRate, channels)
        //if (result != null) {
        // Handle initialization error if needed
        //}
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == REQUEST_AUDIO_PERMISSION &&
                !(grantResults.isNotEmpty()
                            && grantResults[0] == PackageManager.PERMISSION_GRANTED)) {
                showPermissionDeniedDialog()
        }
    }

    private fun showPermissionDeniedDialog() {
        AlertDialog.Builder(this)
            .setTitle("Permission Required")
            .setMessage("This app requires access to the audio output to generate sound. Please grant the permission to continue.")
            .setPositiveButton("Grant Permission") { dialog, _ ->
                dialog.dismiss()
                ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.RECORD_AUDIO), REQUEST_AUDIO_PERMISSION)
            }
            .setNegativeButton("Exit App") { dialog, _ ->
                dialog.dismiss()
                finish()
            }
            .setCancelable(false)
            .show()
    }

    override fun onDestroy() {
        super.onDestroy()
        Mobile_fsg_main.cleanupAudio()
    }
}
