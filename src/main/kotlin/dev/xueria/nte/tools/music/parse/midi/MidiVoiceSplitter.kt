package dev.xueria.nte.tools.music.parse.midi

/**
 * Splits a polyphonic part into monophonic voices.
 *
 * The sheet model stores notes as a plain sequence without absolute time, so a
 * track can only sound one note at a time. Notes are distributed greedily into
 * the first voice that is already silent when they start.
 */
internal object MidiVoiceSplitter {

    fun split(notes: List<MidiNote>): List<List<MidiNote>> {
        val ordered: List<MidiNote> = notes.sortedWith(ONSET_ORDER)
        val voices: MutableList<Voice> = mutableListOf()

        ordered.forEach { note -> add(note, voices) }

        return voices.map { voice -> voice.notes }
    }

    private fun add(note: MidiNote, voices: MutableList<Voice>) {
        val available: Voice? = voices.firstOrNull { voice -> voice.endsAt <= note.startTick }

        if (available != null) {
            available.add(note)
            return
        }

        val voice: Voice = Voice()
        voice.add(note)
        voices += voice
    }

    private class Voice {
        val notes: MutableList<MidiNote> = mutableListOf()
        var endsAt: Long = 0

        fun add(note: MidiNote) {
            notes += note
            endsAt = note.startTick + note.durationTicks
        }
    }

    private val ONSET_ORDER: Comparator<MidiNote> =
        compareBy({ note -> note.startTick }, { note -> note.pitch })

}
