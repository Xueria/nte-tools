package dev.xueria.nte.tools.music.parse.midi

/** Thrown when MIDI data cannot be turned into a [MidiSong]. */
internal class MidiFormatException(
    message: String,
    cause: Throwable? = null
) : RuntimeException(message, cause)

/** A sounding note extracted from one MIDI track. Ticks are `Long`: a sequence may exceed `Int`. */
internal data class MidiNote(
    val track: Int,
    val channel: Int,
    val pitch: Int,
    val startTick: Long,
    val durationTicks: Long
)

/** A `Set Tempo` meta event, in microseconds per quarter note. */
internal data class MidiTempo(val tick: Long, val microsecondsPerQuarter: Int)

/** A `Time Signature` meta event with the denominator already expanded. */
internal data class MidiTimeSignature(val tick: Long, val top: Int, val bottom: Int)

/** A `Key Signature` meta event; [sharps] is negative for flat keys. */
internal data class MidiKeySignature(val tick: Long, val sharps: Int, val minor: Boolean)

/** The content of one MIDI track. */
internal data class MidiTrack(
    val index: Int,
    val name: String?,
    val tempos: List<MidiTempo>,
    val timeSignatures: List<MidiTimeSignature>,
    val keySignatures: List<MidiKeySignature>,
    val notes: List<MidiNote>
)

/** A complete MIDI file, still expressed in ticks. */
internal data class MidiSong(
    val name: String?,
    val divisionType: Float,
    val resolution: Int,
    val microsecondsPerQuarter: Int,
    val timeSignature: MidiTimeSignature?,
    val keySignature: MidiKeySignature?,
    val notes: List<MidiNote>
)

private const val MICROSECONDS_PER_SECOND: Float = 1_000_000f

/** Converts MIDI ticks into quarter-note beats. */
internal sealed interface MidiTickConverter {

    fun beatsOf(ticks: Long): Float

    /** Pulses per quarter note, the common case. */
    class PerQuarter(private val ticksPerQuarter: Int) : MidiTickConverter {

        override fun beatsOf(ticks: Long): Float = ticks.toFloat() / ticksPerQuarter

    }

    /**
     * SMPTE frames. Absolute time is mapped back to beats with the first tempo,
     * because a constant tempo is all the sheet model can express.
     */
    class Smpte(
        private val ticksPerSecond: Float,
        private val microsecondsPerQuarter: Int
    ) : MidiTickConverter {

        override fun beatsOf(ticks: Long): Float =
            ticks.toFloat() * microsecondsPerQuarter / (MICROSECONDS_PER_SECOND * ticksPerSecond)

    }

    companion object {

        /** `Sequence.PPQ` equals zero, every SMPTE division type is a frame rate. */
        private const val PULSES_PER_QUARTER: Float = 0.0f

        fun of(divisionType: Float, resolution: Int, microsecondsPerQuarter: Int): MidiTickConverter {
            if (resolution <= 0) throw MidiFormatException("Invalid MIDI resolution: $resolution")

            if (divisionType == PULSES_PER_QUARTER) return PerQuarter(resolution)

            return Smpte(
                ticksPerSecond = divisionType * resolution,
                microsecondsPerQuarter = microsecondsPerQuarter
            )
        }

    }

}
