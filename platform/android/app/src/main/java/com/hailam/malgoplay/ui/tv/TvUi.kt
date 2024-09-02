package com.hailam.malgoplay.ui.tv

import androidx.compose.foundation.layout.*
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.ExposedDropdownMenuBox
import androidx.compose.material3.ExposedDropdownMenuDefaults
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.TextField
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.unit.dp
import androidx.tv.material3.*
import androidx.tv.material3.MaterialTheme
import com.hailam.malgoplay.MainViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TvUI(viewModel: MainViewModel) {
    var maxFrequency by remember { mutableStateOf(MainViewModel.DEFAULT_MAX_FREQUENCY.toString()) }
    var minFrequency by remember { mutableStateOf(MainViewModel.DEFAULT_MIN_FREQUENCY.toString()) }
    var amplitude by remember { mutableStateOf(MainViewModel.DEFAULT_AMPLITUDE.toString()) }
    var sweepRate by remember { mutableStateOf(MainViewModel.DEFAULT_SWEEP_RATE.toString()) }
    var channels by remember { mutableStateOf(MainViewModel.DEFAULT_CHANNELS.toString()) }
    var selectedMode by remember { mutableStateOf("Sine") }
    var isPlaying by remember { mutableStateOf(false) }

    val modes = listOf("Sine", "Linear", "Triangle", "Random", "Exponential", "Logarithmic", "Square")
    val focusRequester = remember { FocusRequester() }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(32.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Column(modifier = Modifier.weight(1f)) {
                OutlinedTextField(
                    value = maxFrequency,
                    onValueChange = { maxFrequency = it },
                    label = { Text("Max Frequency (Hz)") },
                    modifier = Modifier.fillMaxWidth()
                )
                OutlinedTextField(
                    value = minFrequency,
                    onValueChange = { minFrequency = it },
                    label = { Text("Min Frequency (Hz)") },
                    modifier = Modifier.fillMaxWidth()
                )
                OutlinedTextField(
                    value = amplitude,
                    onValueChange = { amplitude = it },
                    label = { Text("Amplitude") },
                    modifier = Modifier.fillMaxWidth()
                )
            }
            Column(modifier = Modifier.weight(1f)) {
                OutlinedTextField(
                    value = sweepRate,
                    onValueChange = { sweepRate = it },
                    label = { Text("Sweep Rate (Hz/s)") },
                    modifier = Modifier.fillMaxWidth()
                )
                OutlinedTextField(
                    value = channels,
                    onValueChange = { channels = it },
                    label = { Text("Channels") },
                    modifier = Modifier.fillMaxWidth()
                )
                ExposedDropdownMenuBox(
                    expanded = false,
                    onExpandedChange = { },
                ) {
                    TextField(
                        value = selectedMode,
                        onValueChange = { },
                        readOnly = true,
                        label = { Text("Sweep Mode") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = false) },
                        modifier = Modifier.menuAnchor().fillMaxWidth()
                    )
                    ExposedDropdownMenu(
                        expanded = false,
                        onDismissRequest = { },
                    ) {
                        modes.forEach { mode ->
                            DropdownMenuItem(
                                text = { Text(mode) },
                                onClick = { selectedMode = mode }
                            )
                        }
                    }
                }
            }
        }

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
            modifier = Modifier
                .align(Alignment.CenterHorizontally)
                .focusRequester(focusRequester),
        ) {
            Text(if (isPlaying) "Stop" else "Play")
        }
    }

    LaunchedEffect(Unit) {
        focusRequester.requestFocus()
    }
}