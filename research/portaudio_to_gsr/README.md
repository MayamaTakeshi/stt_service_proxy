### portaudio_to_gsr

This is golang code showing how to get audio from microphone and send to to google speech-to-text service to do continuous recognition.

## Preparation

Wet your credentials by doing:
```
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials/file
```
Then build and start the app:
```
go build
./portaudio_to_gsr
```

If the app fails to start, you might need to install some portaudio packages.

I don't remember which ones exactly but each one of these till the app starts successfully:

```
sudo apt install libportaudio2

sudo apt install portaudio19-dev

sudo app install libportaudiocpp0
```

## Sample execution

You might see some portaudio related error messages but as long as the app stays alive, it should be OK:
```
takeshi@takeshi-desktop:portaudio_to_gsr$ ./portaudio_to_gsr 
ALSA lib pcm_dmix.c:1032:(snd_pcm_dmix_open) unable to open slave
ALSA lib pcm.c:2664:(snd_pcm_open_noupdate) Unknown PCM cards.pcm.rear
ALSA lib pcm.c:2664:(snd_pcm_open_noupdate) Unknown PCM cards.pcm.center_lfe
ALSA lib pcm.c:2664:(snd_pcm_open_noupdate) Unknown PCM cards.pcm.side
ALSA lib pcm_route.c:877:(find_matching_chmap) Found no matching channel map
connect(2) call to /dev/shm/jack-1000/default/jack_0 failed (err=No such file or directory)
attempt to connect to server failed
connect(2) call to /dev/shm/jack-1000/default/jack_0 failed (err=No such file or directory)
attempt to connect to server failed
ALSA lib pcm_oss.c:397:(_snd_pcm_oss_open) Cannot open device /dev/dsp
ALSA lib pcm_oss.c:397:(_snd_pcm_oss_open) Cannot open device /dev/dsp
ALSA lib confmisc.c:160:(snd_config_get_card) Invalid field card
ALSA lib pcm_usb_stream.c:482:(_snd_pcm_usb_stream_open) Invalid card 'card'
ALSA lib confmisc.c:160:(snd_config_get_card) Invalid field card
ALSA lib pcm_usb_stream.c:482:(_snd_pcm_usb_stream_open) Invalid card 'card'
ALSA lib pcm_dmix.c:1032:(snd_pcm_dmix_open) unable to open slave
connect(2) call to /dev/shm/jack-1000/default/jack_0 failed (err=No such file or directory)
attempt to connect to server failed
Transcription: hello
Transcription:  good morning
Transcription:  how are you
Transcription:  call James
Transcription:  it was working
Transcription:  it was working
Transcription:  it's fine
Transcription:  very good
Transcription:  excellent
Transcription:  oh it is it is going to work
Transcription:  oh this is going to work
Transcription:  this is going to work
Transcription:  I don't know
Transcription:  I'm sure
Transcription:  trust me
Transcription:  trust your culture Mary
Transcription:  trust her to marry
Transcription:  transfer to Mary
Transcription:  transfer to Mary
Transcription:  transfer to Steve
Transcription:  transfer call to Steve
Transcription:  transfer call to Mary
Transcription:  transfer to Mary
Transcription:  search Mary
Transcription:  search teeth
Transcription:  search David
Transcription:  search Steve
Transcription:  search John
Transcription:  search John
Transcription:  nice nice
```

