package com.hailam.malgoplay.ui.common

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.ColorScheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.tv.material3.ExperimentalTvMaterial3Api
import androidx.tv.material3.darkColorScheme as tvDarkColorScheme
import androidx.tv.material3.lightColorScheme as tvLightColorScheme

private val calmBlue = Color(0xFF4A86E8)
private val softGreen = Color(0xFF81C784)

private val LightColors = lightColorScheme(
    primary = calmBlue,
    onPrimary = Color.White,
    secondary = softGreen,
    onSecondary = Color.Black,
    background = Color.White,
    onBackground = Color.Black,
    surface = Color.White,
    onSurface = Color.Black,
    surfaceVariant = Color(0xFFF0F0F0),
    onSurfaceVariant = Color.Black,
)

private val DarkColors = darkColorScheme(
    primary = calmBlue,
    onPrimary = Color.White,
    secondary = softGreen,
    onSecondary = Color.Black,
    background = Color(0xFF121212),
    onBackground = Color.White,
    surface = Color(0xFF1E1E1E),
    onSurface = Color.White,
    surfaceVariant = Color(0xFF2D2D2D),
    onSurfaceVariant = Color.White,
)

@OptIn(ExperimentalTvMaterial3Api::class)
private val TvLightColors = tvLightColorScheme(
    primary = calmBlue,
    onPrimary = Color.White,
    secondary = softGreen,
    onSecondary = Color.Black,
    background = Color(0xFFF5F5F5),
    onBackground = Color.Black,
    surface = Color.White,
    onSurface = Color.Black,
    surfaceVariant = Color(0xFFE0E0E0),
    onSurfaceVariant = Color.Black,
)

@OptIn(ExperimentalTvMaterial3Api::class)
private val TvDarkColors = tvDarkColorScheme(
    primary = calmBlue,
    onPrimary = Color.White,
    secondary = softGreen,
    onSecondary = Color.Black,
    background = Color(0xFF121212),
    onBackground = Color.White,
    surface = Color(0xFF1E1E1E),
    onSurface = Color.White,
    surfaceVariant = Color(0xFF2D2D2D),
    onSurfaceVariant = Color.White,
)

@Composable
fun MainTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    isTvTheme: Boolean = false,
    content: @Composable () -> Unit
) {
    val colors = when {
        isTvTheme && darkTheme -> TvDarkColors
        isTvTheme && !darkTheme -> TvLightColors
        !isTvTheme && darkTheme -> DarkColors
        else -> LightColors
    }

    MaterialTheme(
        colorScheme = colors as ColorScheme,
        typography = MaterialTheme.typography,
        content = content
    )
}