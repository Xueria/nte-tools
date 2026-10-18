package dev.xueria.nte.tools.json

import kotlinx.serialization.json.Json

val JsonFileExtension = "json"

val GlobalJson = Json {
    ignoreUnknownKeys = true
    prettyPrint = true
}