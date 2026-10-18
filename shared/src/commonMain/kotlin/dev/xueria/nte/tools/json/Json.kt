package dev.xueria.nte.tools.json

import kotlinx.serialization.json.Json

val GlobalJson = Json {
    ignoreUnknownKeys = true
    prettyPrint = true
}