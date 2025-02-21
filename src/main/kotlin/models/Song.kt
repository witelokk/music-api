package models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Song(
    val id: String,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    @SerialName("is_favorite") val isFavorite: Boolean,
    @SerialName("duration_seconds") val durationSeconds: Int,
    val artists: List<ShortArtist>,
    @SerialName("stream_url") val streamUrl: String,
)