package com.hailam.malgoplay.ui.mobile

import androidx.compose.animation.*
import androidx.compose.foundation.ScrollState
import androidx.compose.foundation.gestures.Orientation
import androidx.compose.foundation.gestures.scrollable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.scale
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import com.hailam.malgoplay.MainViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MobileUI(viewModel: MainViewModel) {
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
        "Exponential", "Logarithmic", "Square"
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

        Spacer(modifier = Modifier.height(24.dp))

        Column(
            verticalArrangement = Arrangement.Bottom,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Button(
                modifier = Modifier.scale(1.3F),
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
                )
            }
        }
    }
}