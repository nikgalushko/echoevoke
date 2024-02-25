## TODO 
- [X] github actions to start tests
- [ ] new channel register
- [X] scrapper 
- [X] images downloader
- [ ] sqlite storage
- [ ] landing page
- [ ] `channelID` in registry must be a valid and exist channel

## Images downloader
- [ ] before download an image do HEAD request and check Etag; if we already download the image with same etag do not download it again

## Scrapper
- [ ] if the channel is not registered then skip it

## Storage
- [ ] use sync.RWMutes in `memstorage`