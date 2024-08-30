package com.hailam.malgoplay

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.widget.Button
import android.widget.EditText
import androidx.appcompat.app.AlertDialog
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import malgoplay_android_main.Malgoplay_android_main

class MainActivity : AppCompatActivity() {

    companion object {
        private const val REQUEST_AUDIO_PERMISSION = 1
        private const val DEFAULT_MAX_FREQUENCY = 1000.0
        private const val DEFAULT_MIN_FREQUENCY = 1000.0
        private const val DEFAULT_AMPLITUDE = 0.5
        private const val DEFAULT_SWEEP_RATE = 1.0
        private const val UPDATE_INTERVAL = 100L
    }

    private lateinit var frequencyUpdateHandler: Handler
    private var shouldRun = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        frequencyUpdateHandler = Handler(Looper.getMainLooper())

        if (ContextCompat.checkSelfPermission(this, Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.RECORD_AUDIO), REQUEST_AUDIO_PERMISSION)
        } else {
            initAudioDevice()
        }

        // Get references to the UI elements
        val frequencyInput: EditText = findViewById(R.id.frequency_input)
        val amplitudeInput: EditText = findViewById(R.id.amplitude_input)
        val minFrequencyInput: EditText = findViewById(R.id.min_frequency_input)
        val sweepRateInput: EditText = findViewById(R.id.sweep_rate_input)
        val playButton: Button = findViewById(R.id.play_button)
        val stopButton: Button = findViewById(R.id.stop_button)

        // Set default values
        frequencyInput.setText(DEFAULT_MAX_FREQUENCY.toString())
        amplitudeInput.setText(DEFAULT_AMPLITUDE.toString())
        minFrequencyInput.setText(DEFAULT_MIN_FREQUENCY.toString())
        sweepRateInput.setText(DEFAULT_SWEEP_RATE.toString())

        // Set an onClick listener for the Play button
        playButton.setOnClickListener {
            val maxFrequency = frequencyInput.text.toString().toDoubleOrNull() ?: DEFAULT_MAX_FREQUENCY
            val amplitude = amplitudeInput.text.toString().toDoubleOrNull() ?: DEFAULT_AMPLITUDE
            val minFrequency = minFrequencyInput.text.toString().toDoubleOrNull() ?: DEFAULT_MIN_FREQUENCY
            val sweepRate = sweepRateInput.text.toString().toDoubleOrNull() ?: DEFAULT_SWEEP_RATE

            Malgoplay_android_main.setVolume(1.0)  // Ensure volume is reset to 1.0 before playing
            Malgoplay_android_main.setFrequency(maxFrequency)
            Malgoplay_android_main.setMinFrequency(minFrequency)
            Malgoplay_android_main.setAmplitude(amplitude)
            Malgoplay_android_main.setSweepRate(sweepRate)

            shouldRun = true
            Malgoplay_android_main.startDevice()
            startFrequencyUpdateLoop()
        }

        // Set an onClick listener for the Stop button
        stopButton.setOnClickListener {
            shouldRun = false
            Malgoplay_android_main.stopDevice()
            frequencyUpdateHandler.removeCallbacksAndMessages(null)
        }
    }

    private fun startFrequencyUpdateLoop() {
        frequencyUpdateHandler.post(object : Runnable {
            override fun run() {
                if (shouldRun) {
                    Malgoplay_android_main.updateFrequency()

                    
                    frequencyUpdateHandler.postDelayed(this, UPDATE_INTERVAL)
                }
            }
        })
    }

    private fun initAudioDevice() {
        Malgoplay_android_main.initDevice()
    }

    override fun onRequestPermissionsResult(requestCode: Int, permissions: Array<out String>, grantResults: IntArray) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == REQUEST_AUDIO_PERMISSION) {
            if (grantResults.isNotEmpty() && grantResults[0] == PackageManager.PERMISSION_GRANTED) {
                initAudioDevice()
            } else {
                showPermissionDeniedDialog()
            }
        }
    }

    private fun showPermissionDeniedDialog() {
        AlertDialog.Builder(this)
            .setTitle("Permission Required")
            .setMessage("This app requires access to the microphone to generate and analyze sound. Please grant the permission to continue.")
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
        shouldRun = false
        frequencyUpdateHandler.removeCallbacksAndMessages(null)
        Malgoplay_android_main.cleanupDevice()
    }
}
