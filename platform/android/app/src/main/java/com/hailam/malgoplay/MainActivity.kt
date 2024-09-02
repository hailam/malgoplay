package com.hailam.malgoplay

import android.Manifest
import android.content.pm.PackageManager
import android.os.Bundle
import android.view.WindowManager
import androidx.core.app.ActivityCompat
import androidx.appcompat.app.AlertDialog
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.activity.viewModels
import androidx.appcompat.app.AppCompatActivity
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.ScrollState
import androidx.compose.foundation.gestures.Orientation
import androidx.compose.foundation.gestures.scrollable
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import kotlinx.coroutines.*
import mobile_fsg_main.Mobile_fsg_main

class MainActivity : AppCompatActivity() {

    private val viewModel: MainViewModel by viewModels()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Request permission to use audio
        val requestPermissionLauncher = registerForActivityResult(
            ActivityResultContracts.RequestPermission()
        ) { isGranted: Boolean ->
            if (!isGranted) {
                showPermissionDeniedDialog()
            }
        }

        if (ContextCompat.checkSelfPermission(this, Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            requestPermissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
        }

        window.setFlags(
            WindowManager.LayoutParams.FLAG_FULLSCREEN,
            WindowManager.LayoutParams.FLAG_FULLSCREEN
        )

        setContent {
            MainTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    MalgoplayUI(viewModel)
                }
            }
        }
    }

    private fun showPermissionDeniedDialog() {
        AlertDialog.Builder(this)
            .setTitle("Permission Required")
            .setMessage("This app requires access to the microphone to detect the frequency of the audio signal.")
            .setPositiveButton("Grant Permission") { dialog, _ ->
                dialog.dismiss()
                ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.RECORD_AUDIO), 1)
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
        CoroutineScope(Dispatchers.IO).launch {
            Mobile_fsg_main.cleanupAudio()
        }
    }
}


@Composable
fun MainTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    content: @Composable () -> Unit
) {
    val calmBlue = Color(0xFF4A86E8)

    val colors = if (darkTheme) {
        darkColorScheme(
            primary = calmBlue,
            onPrimary = Color.White,
            secondary = Color(0xFF81C784), // A softer green
            onSecondary = Color.Black,
            background = Color(0xFF121212),
            onBackground = Color.White,
            surface = Color(0xFF1E1E1E),
            onSurface = Color.White,
            surfaceVariant = Color(0xFF2D2D2D),
            onSurfaceVariant = Color.White,
        )
    } else {
        lightColorScheme(
            primary = calmBlue,
            onPrimary = Color.White,
            secondary = Color(0xFF81C784), // A softer green
            onSecondary = Color.Black,
            background = Color.White,
            onBackground = Color.Black,
            surface = Color.White,
            onSurface = Color.Black,
            surfaceVariant = Color(0xFFF0F0F0),
            onSurfaceVariant = Color.Black,
        )
    }

    MaterialTheme(
        colorScheme = colors,
        content = content
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MalgoplayUI(viewModel: MainViewModel) {
    var maxFrequency by remember { mutableStateOf(MainViewModel.DEFAULT_MAX_FREQUENCY.toString()) }
    var minFrequency by remember { mutableStateOf(MainViewModel.DEFAULT_MIN_FREQUENCY.toString()) }
    var amplitude by remember { mutableStateOf(MainViewModel.DEFAULT_AMPLITUDE.toString()) }
    var sweepRate by remember { mutableStateOf(MainViewModel.DEFAULT_SWEEP_RATE.toString()) }
    var channels by remember { mutableStateOf(MainViewModel.DEFAULT_CHANNELS.toString()) }
    var selectedMode by remember { mutableStateOf("Sine") }
    var isPlaying by remember { mutableStateOf(false) }

    var modeListExpanded by remember { mutableStateOf(false) }
    val modes = listOf(
        "Sine", "Linear", "Triangle", "Random",
        "Exponential", "Logarithmic", "Square",
        /*"Sawtooth", Sawtooth is just Linear*/
    )

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
            .scrollable(ScrollState(0), Orientation.Vertical),
        verticalArrangement = Arrangement.Center,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {

        AnimatedVisibility(
            visible = !isPlaying,
            enter = fadeIn() + expandVertically(),
            exit = fadeOut() + shrinkVertically()
        ) {
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                ExposedDropdownMenuBox(
                    expanded = modeListExpanded,
                    onExpandedChange = { modeListExpanded = !modeListExpanded }
                ) {
                    TextField(
                        value = selectedMode,
                        onValueChange = { selectedMode = it },
                        label = { Text("Select Mode") },
                        readOnly = true,
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = modeListExpanded) },
                        colors = TextFieldDefaults.colors(),
                        modifier = Modifier.menuAnchor()
                    )
                    ExposedDropdownMenu(
                        expanded = modeListExpanded,
                        onDismissRequest = { modeListExpanded = false }
                    ) {
                        modes.forEach { mode ->
                            DropdownMenuItem(
                                text = { Text(text = mode) },
                                onClick = {
                                    selectedMode = mode
                                    modeListExpanded = false
                                }
                            )
                        }
                    }
                }

                OutlinedTextField(
                    value = maxFrequency,
                    onValueChange = { maxFrequency = it },
                    label = { Text("Enter Max Frequency (Hz)") },
                    keyboardOptions = KeyboardOptions.Default.copy(keyboardType = KeyboardType.Number)
                )

                OutlinedTextField(
                    value = amplitude,
                    onValueChange = { amplitude = it },
                    label = { Text("Enter Amplitude") },
                    keyboardOptions = KeyboardOptions.Default.copy(keyboardType = KeyboardType.Number)
                )

                OutlinedTextField(
                    value = minFrequency,
                    onValueChange = { minFrequency = it },
                    label = { Text("Enter Min Frequency (Hz)") },
                    keyboardOptions = KeyboardOptions.Default.copy(keyboardType = KeyboardType.Number)
                )

                OutlinedTextField(
                    value = sweepRate,
                    onValueChange = { sweepRate = it },
                    label = { Text("Enter Sweep Rate (Hz per second)") },
                    keyboardOptions = KeyboardOptions.Default.copy(keyboardType = KeyboardType.Number)
                )

                OutlinedTextField(
                    value = channels,
                    onValueChange = { channels = it },
                    label = { Text("Channels") },
                    keyboardOptions = KeyboardOptions.Default.copy(keyboardType = KeyboardType.Number)
                )
            }
        }

        Spacer(modifier = Modifier.height(16.dp))

        Button(
            onClick = {
                if (!isPlaying) {
                    viewModel.startPlaying(
                        minFrequency.toDouble(),
                        maxFrequency.toDouble(),
                        amplitude.toDouble(),
                        sweepRate.toDouble(),
                        channels.toLong(),
                        selectedMode
                    )
                } else {
                    viewModel.stopPlaying()
                }
                isPlaying = !isPlaying
            },
            colors = ButtonDefaults.buttonColors(
                containerColor = if (isPlaying) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary,
                contentColor = MaterialTheme.colorScheme.onPrimary
            ),
            shape = MaterialTheme.shapes.medium
        ) {
            Text(
                text = if (isPlaying) "Stop" else "Play",
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = FontWeight.Bold
            )
        }
    }
}
