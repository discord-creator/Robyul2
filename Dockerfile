FROM sn0w/go-ci

# Install deps
RUN apk add --update curl wget python py-setuptools ffmpeg gnupg

# Get youtube-dl
RUN wget https://yt-dl.org/downloads/latest/youtube-dl.sig -O /tmp/youtube-dl.sig
RUN curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /usr/bin/youtube-dl

# Verify the release
RUN gpg --keyserver pool.sks-keyservers.net --recv-keys DB4B54CBA4826A18
RUN gpg --keyserver pool.sks-keyservers.net --recv-keys 2C393E0F18A9236D
RUN gpg --verify /tmp/youtube-dl.sig /usr/bin/youtube-dl
RUN chmod a+rx /usr/bin/youtube-dl
RUN rm /tmp/youtube-dl.sig

# Expose karen's api
EXPOSE 1337

# Define volumes
VOLUME /karen
WORKDIR /karen

# Define upstart
ENTRYPOINT /karen/karen
