package models

import com.witelokk.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.UUID

@Serializable
data class ShortArtist(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("avatar_url") val avatarUrl: String?,
)