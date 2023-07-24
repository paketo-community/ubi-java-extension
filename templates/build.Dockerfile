ARG base_image
FROM ${base_image}

USER root

ARG build_id=0
RUN echo ${build_id}

RUN microdnf --setopt=install_weak_deps=0 --setopt=tsflags=nodocs install -y {{.PACKAGES}} && microdnf clean all

RUN echo "{{.JAVA_VERSION}}" > /bpi.paketo.ubi.java.version
RUN echo "{{.JAVA_EXTENSION_HELPERS}}" > /bpi.paketo.ubi.java.helpers

USER {{.CNB_USER_ID}}:{{.CNB_GROUP_ID}}