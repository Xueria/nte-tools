import org.jetbrains.compose.desktop.application.dsl.TargetFormat

plugins {
    alias(libs.plugins.kotlin.multiplatform)
    alias(libs.plugins.compose.multiplatform)
    alias(libs.plugins.compose.compiler)
}

//dependencies {
//    implementation(project(":shared"))
//
//    implementation(compose.desktop.currentOs)
//    implementation(libs.kotlinx.coroutinesSwing)
//
//    implementation(libs.compose.uiToolingPreview)
//}

kotlin {
    jvm()

    jvmToolchain(21)

    sourceSets {
        jvmMain.dependencies {
            implementation(project(":shared"))

            implementation(libs.compose.components.resources)
            implementation(compose.desktop.currentOs)
            implementation(libs.kotlinx.coroutines.swing)

            implementation(libs.compose.ui.tooling.preview)
        }
    }
}

compose.desktop {
    application {
        mainClass = "dev.xueria.nte.tools.MainKt"

        nativeDistributions {
            targetFormats(TargetFormat.Exe, TargetFormat.Msi)

            packageName = "nte-music"
            packageVersion = "1.0.1"

            val iconRoot = project.file("icons")

            windows {
                iconFile = iconRoot.resolve("windows.ico")
            }
        }
    }
}