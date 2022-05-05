package resource

import (
	"io"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

type Audio struct {
	Path string

	// Volume adjust how loud this sound will be.
	// The default value of 0 means "unadjusted".
	// Value greated than 0 increases the volume, negative values decrease it.
	// This setting accepts values in [-1, 1] range, where -1 mutes the sound
	// while 1 makes it as loud as possible.
	Volume float64
}

type AudioID int

type AudioRegistry struct {
	mapping map[AudioID]Audio
}

func (r *AudioRegistry) Set(id AudioID, info Audio) {
	r.mapping[id] = info
}

type AudioSystem struct {
	loader *Loader

	currentMusic *audioResource

	audioContext *audio.Context
	resources    map[AudioID]*audioResource
}

type audioResource struct {
	player *audio.Player
	volume float64
}

func (sys *AudioSystem) Init(l *Loader) {
	sys.loader = l
	sys.audioContext = audio.NewContext(32000)
	sys.resources = make(map[AudioID]*audioResource)
}

func (sys *AudioSystem) DecodeWAV(r io.Reader) (*wav.Stream, error) {
	return wav.Decode(sys.audioContext, r)
}

func (sys *AudioSystem) DecodeOGG(r io.Reader) (*vorbis.Stream, error) {
	return vorbis.Decode(sys.audioContext, r)
}

func (sys *AudioSystem) getOGGResource(id AudioID) *audioResource {
	resource, ok := sys.resources[id]
	if ok {
		return resource
	}
	stream := sys.loader.LoadOGG(id)
	oggInfo := sys.loader.GetAudioInfo(id)
	loopedStream := audio.NewInfiniteLoop(stream, stream.Length())
	player, err := audio.NewPlayer(sys.audioContext, loopedStream)
	if err != nil {
		panic(err.Error())
	}
	volume := (oggInfo.Volume / 2) + 0.5
	resource = &audioResource{
		player: player,
		volume: volume,
	}
	sys.resources[id] = resource
	return resource
}

func (sys *AudioSystem) PauseCurrentMusic() {
	if sys.currentMusic == nil {
		return
	}
	sys.currentMusic.player.Pause()
}

func (sys *AudioSystem) ContinueCurrentMusic() {
	if sys.currentMusic == nil || sys.currentMusic.player.IsPlaying() {
		return
	}
	sys.currentMusic.player.SetVolume(sys.currentMusic.volume)
	sys.currentMusic.player.Play()
}

func (sys *AudioSystem) ContinueMusic(id AudioID) {
	resource := sys.getOGGResource(id)
	if resource.player.IsPlaying() {
		return
	}
	sys.currentMusic = resource
	resource.player.SetVolume(resource.volume)
	resource.player.Play()
}

func (sys *AudioSystem) PlayMusic(id AudioID) {
	resource := sys.getOGGResource(id)
	if sys.currentMusic != nil && resource.player == sys.currentMusic.player {
		return
	}
	sys.currentMusic = resource
	resource.player.SetVolume(resource.volume)
	resource.player.Rewind()
	resource.player.Play()
}

func (sys *AudioSystem) PlaySound(id AudioID) {
	resource, ok := sys.resources[id]
	if !ok {
		stream := sys.loader.LoadWAV(id)
		wavInfo := sys.loader.GetAudioInfo(id)
		player, err := audio.NewPlayer(sys.audioContext, stream)
		if err != nil {
			panic(err.Error())
		}
		volume := (wavInfo.Volume / 2) + 0.5
		resource = &audioResource{
			player: player,
			volume: volume,
		}
		sys.resources[id] = resource
	}
	resource.player.SetVolume(resource.volume)
	resource.player.Rewind()
	resource.player.Play()
}
