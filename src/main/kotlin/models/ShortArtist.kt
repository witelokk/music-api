package models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class ShortArtist(
    val id: String,
    val name: String,
    @SerialName("avatar_url") val avatarUrl: String?,
)