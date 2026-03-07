import groovy.json.JsonOutput
def Params

pipeline {
    agent {
        kubernetes {
            label 'codlocker-assets'
            defaultContainer 'jnlp'
            yamlFile 'KubernetesPod.yaml'
        }
    }

    environment {
       
        COMMIT_FILES = sh(script: 'git show --pretty="" --name-only', returnStdout: true).trim()
        REVISION = sh(script: 'git rev-parse', returnStdout: true).trim()
        DOCKER_DEST = "gururepservice"
        BMAJOR = "1"
        MICRO_SERVICE_NAME = "codlocker-assets"
    }

    stages {
    stage('Build with Kaniko') {
          steps {
            container(name: 'kaniko', shell: '/busybox/sh') {
              withEnv(['PATH+EXTRA=/busybox']) {
                sh '''#!/busybox/sh
                  set -euo pipefail
        
                  test -s "${WORKSPACE}/Dockerfile" || { echo "Dockerfile missing"; exit 2; }
                  [ -s /kaniko/.docker/config.json ] || { echo "Missing /kaniko/.docker/config.json (secret)"; exit 3; }
        
                  # derive values
                  ARTIFACT_ID="${MICRO_SERVICE_NAME}"
                  VERSION="${BMAJOR}.${BUILD_ID}"
                  COMMIT_SHA="${REVISION:-${GIT_COMMIT:-unknown}}"
        
                  DIGEST_FILE="${WORKSPACE}/.kaniko.digest"
                  rm -f "$DIGEST_FILE" || true
        
                  /kaniko/executor \
                    --context "${WORKSPACE}" \
                    --dockerfile "${WORKSPACE}/Dockerfile" \
                    --destination "${DOCKER_DEST}/${MICRO_SERVICE_NAME}:${VERSION}" \
                    --cache=true \
                    --verbosity=info \
                    --digest-file "$DIGEST_FILE" \
                    --build-arg ARTIFACT_ID="${ARTIFACT_ID}" \
                    --build-arg VERSION="${VERSION}" \
                    --build-arg COMMIT_SHA="${COMMIT_SHA}"
        
                  [ -s "$DIGEST_FILE" ] && echo "Pushed digest: $(cat "$DIGEST_FILE")" || true
                '''
              }
            }
          }
        }
        stage('Echo Artifact Registration Payload (dry-run)') {
          steps {
            script {
              def digestPath = "${env.WORKSPACE}/.kaniko.digest"
              def digest = fileExists(digestPath) ? readFile(digestPath).trim() : ""
              def tag = "${DOCKER_DEST}/codlocker-assets:${BMAJOR}.${BUILD_ID}"
        
              def payload = """
              {
                "component_id": "codlocker-assets",
                "artifact_type": "container-image",
                "artifact_name": "gururepservice/codlocker-assets",
                "reference": "${tag}",
                "version": "${BMAJOR}.${BUILD_ID}",
                "digest": "${digest}",
                "ci_system": "jenkins",
                "job_name": "${env.JOB_NAME}",
                "build_number": "${env.BUILD_NUMBER}",
                "git_commit": "${env.GIT_COMMIT ?: ""}"
              }
              """.stripIndent().trim()
        
              echo "=== BEGIN Unify registration payload (dry run) ==="
              echo payload
              echo "=== END Unify registration payload ==="
            }
          }
        }

        stage('Artifact Registration') {
          steps {
            script {
              def digestFile = "${env.WORKSPACE}/.kaniko.digest"
              def digest   = fileExists(digestFile) ? readFile(digestFile).trim() : ""
        
              def fullSha  = env.GIT_COMMIT ?: sh(script: 'git rev-parse HEAD', returnStdout: true).trim()
              def shortSha = fullSha.take(12)
              def buildTs  = sh(script: 'date -u +%Y-%m-%dT%H:%M:%SZ', returnStdout: true).trim()
              def version  = "${BMAJOR}.${BUILD_ID}"
              def reference = "${DOCKER_DEST}/${MICRO_SERVICE_NAME}:${version}"
        
              echo "Registering artifact metadata in Unify..."
              registerBuildArtifactMetadata(
                name: "gururepservice/${MICRO_SERVICE_NAME}",
                url: "${reference}",
                version: version,
                digest: digest,
                label: "build=cbci,version=${version},sha=${shortSha},d.prod=true",
                type: "docker"
              )
        
              // Optional: make it easy to see in Jenkins UI + keep a file artifact
              currentBuild.description = "v${version} • ${shortSha}"
              writeJSON file: 'unify-metadata.json', json: [
                component      : "gururepservice/${MICRO_SERVICE_NAME}",
                version        : version,
                sha_full       : fullSha,
                sha_short      : shortSha,
                digest         : digest,
                reference      : reference,
                timestamp_utc  : buildTs,
                ci_system      : 'jenkins',
                job_name       : env.JOB_NAME,
                build_number   : env.BUILD_NUMBER
              ], pretty: 4
              archiveArtifacts artifacts: 'unify-metadata.json', onlyIfSuccessful: true
            }
          }
        }
                
        
    }
}
