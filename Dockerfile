# Copyright 2022 Richard Kosegi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.22 as builder

WORKDIR /build
COPY . /build

RUN make build-local

FROM cgr.dev/chainguard/static:latest
ARG VERSION
ARG BUILD_DATE
ARG GIT_COMMIT
COPY --from=builder /build/exporter /

LABEL org.opencontainers.image.url="https://github.com/rkosegi/tuya-smartplug-exporter" \
      org.opencontainers.image.documentation="https://github.com/rkosegi/tuya-smartplug-exporter/blob/main/README.md" \
      org.opencontainers.image.source="https://github.com/rkosegi/tuya-smartplug-exporter.git" \
      org.opencontainers.image.title="Tuya smartplug exporter" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="rkosegi" \
      org.opencontainers.image.description="Prometheus exporter for Tuya-based smartplug devices." \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${GIT_COMMIT}" \
      org.opencontainers.image.version="${VERSION}"

USER 65532:65532

WORKDIR /
ENTRYPOINT ["/exporter"]

EXPOSE 9999
