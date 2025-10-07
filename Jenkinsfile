pipeline {
    agent any

    environment {
        APP_NAME = 'go-dev-dashboard'
        DOCKER_REGISTRY = 'yourregistry.com'
        DEPLOY_SERVER = 'devops@yourserver'
        SSH_CREDENTIALS_ID = 'jenkins-ssh-key'
        APP_PORT = '8090'
        TARGET_URL = "http://yourserver:${APP_PORT}"
        TRIVY_VERSION = '0.54.1'
        ZAP_VERSION = '2.15.0'
    }

    stages {

        // === 1. Checkout Code ===
        stage('Checkout Code') {
            steps {
                git(branch: 'master', url: 'https://github.com/laurentmd5/sample-app.git', credentialsId: 'github-token')
                sh '''
                echo "📦 Repository: https://github.com/laurentmd5/sample-app.git"
                echo "📝 Branch: master"
                find . -name "*.go" -type f
                '''
            }
        }

        // === 2. Setup Tools ===
        stage('Setup Environment') {
            steps {
                sh '''
                echo "🔧 Vérification de l'environnement..."
                go version || echo "⚠️ Go non installé"
                docker --version || echo "⚠️ Docker non disponible"

                echo "📥 Installation de Trivy et gosec..."
                which gosec || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
                if ! which trivy; then
                    wget -q https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.deb
                    sudo dpkg -i trivy_${TRIVY_VERSION}_Linux-64bit.deb
                fi
                '''
            }
        }

        // === 3. Build Go Application ===
        stage('Build Go Application') {
            steps {
                sh '''
                echo "🏗️ Construction de l'application Go..."
                if [ ! -f "go.mod" ]; then go mod init hello-app; fi
                go mod tidy
                go build -v -o ${APP_NAME} .
                chmod +x ${APP_NAME}
                '''
            }
        }

        // === 4. Static Code Analysis ===
        stage('Static Analysis (Gosec)') {
            steps {
                sh '''
                echo "🔍 Analyse Statique avec gosec..."
                mkdir -p security-reports
                gosec -fmt=html -out=security-reports/gosec-report.html ./... || true
                go vet ./... > security-reports/govet-output.txt || true
                '''
            }
            post {
                always {
                    publishHTML([
                        reportDir: 'security-reports',
                        reportFiles: 'gosec-report.html',
                        reportName: 'Analyse Statique - Gosec'
                    ])
                }
            }
        }

        // === 5. Unit Tests + Coverage ===
        stage('Unit Tests & Coverage') {
            steps {
                sh '''
                echo "🧪 Tests unitaires..."
                mkdir -p test-reports
                go test -v -coverprofile=test-reports/coverage.out ./... | tee test-reports/test.log
                go tool cover -html=test-reports/coverage.out -o test-reports/coverage.html
                '''
            }
            post {
                always {
                    publishHTML([
                        reportDir: 'test-reports',
                        reportFiles: 'coverage.html',
                        reportName: 'Couverture des Tests'
                    ])
                }
            }
        }

        // === 6. Build Docker Image ===
        stage('Build Docker Image') {
            steps {
                sh '''
                echo "🐳 Construction de l'image Docker..."
                docker build -t ${DOCKER_REGISTRY}/${APP_NAME}:${BUILD_NUMBER} .
                docker tag ${DOCKER_REGISTRY}/${APP_NAME}:${BUILD_NUMBER} ${DOCKER_REGISTRY}/${APP_NAME}:latest
                docker images | grep ${APP_NAME}
                '''
            }
        }

        // === 7. Trivy Container Scan ===
        stage('Container Scan (Trivy)') {
            steps {
                sh '''
                echo "🔒 Scan de l'image Docker..."
                mkdir -p trivy-reports
                trivy image --format template --template "@contrib/html.tpl" -o trivy-reports/container-scan.html ${DOCKER_REGISTRY}/${APP_NAME}:latest
                '''
            }
            post {
                always {
                    publishHTML([
                        reportDir: 'trivy-reports',
                        reportFiles: 'container-scan.html',
                        reportName: 'Scan Sécurité - Trivy'
                    ])
                }
            }
        }

        // === 8. Deploy to Remote Server ===
        stage('Deploy to Ubuntu Server') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(credentialsId: "${SSH_CREDENTIALS_ID}", keyFileVariable: 'SSH_KEY')]) {
                        sh """
                        echo "🚀 Déploiement distant..."
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} '
                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            docker run -d --name ${APP_NAME} -p ${APP_PORT}:${APP_PORT} ${DOCKER_REGISTRY}/${APP_NAME}:latest
                        '
                        """
                    }
                }
            }
        }

        // === 9. OWASP ZAP Dynamic Scan (via Docker) ===
        stage('OWASP ZAP Scan (Docker)') {
            steps {
                sh '''
                echo "🛡️ Scan OWASP ZAP via Docker..."
                mkdir -p zap-reports

                docker run --rm -v $(pwd)/zap-reports:/zap/wrk:rw \
                    owasp/zap2docker-stable zap-baseline.py \
                    -t ${TARGET_URL} \
                    -r zap-report.html \
                    -x zap-report.xml \
                    -J zap-report.json \
                    -I -m 10

                echo "✅ Rapport ZAP généré"
                '''
            }
            post {
                always {
                    publishHTML([
                        reportDir: 'zap-reports',
                        reportFiles: 'zap-report.html',
                        reportName: 'OWASP ZAP - Analyse Dynamique'
                    ])
                    archiveArtifacts artifacts: 'zap-reports/*', fingerprint: true
                }
            }
        }

        // === 10. Environment Security Audit (Lynis) ===
        stage('Environment Audit (Lynis)') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(credentialsId: "${SSH_CREDENTIALS_ID}", keyFileVariable: 'SSH_KEY')]) {
                        sh """
                        echo "🏢 Audit de sécurité de l'environnement..."
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} '
                            sudo apt install -y lynis
                            sudo lynis audit system --quick | tee /tmp/lynis-audit.txt
                        '
                        scp -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER}:/tmp/lynis-audit.txt lynis-reports/ || echo 'Aucun rapport Lynis'
                        """
                    }
                }
            }
        }

        // === 11. Security Summary ===
        stage('Security Summary') {
            steps {
                sh '''
                echo "📊 Génération du résumé de sécurité..."
                mkdir -p security-summary
                cat > security-summary/report.md <<EOF
                # Rapport de Sécurité - Build ${BUILD_NUMBER}

                🔐 **Application**: ${APP_NAME}  
                🌐 **URL**: ${TARGET_URL}  
                📅 **Date**: $(date)

                **Modules d'analyse exécutés :**
                - Gosec : ✔️
                - Trivy : ✔️
                - OWASP ZAP : ✔️
                - Lynis : ✔️
                - Tests unitaires : ✔️
                EOF
                '''
            }
            post {
                always {
                    archiveArtifacts artifacts: 'security-summary/**', fingerprint: true
                }
            }
        }

    } // end stages

    post {
        success {
            echo "🎉 Pipeline DevSecOps terminé avec succès !"
        }
        failure {
            echo "❌ Échec du pipeline - vérifiez les rapports."
        }
        always {
            sh 'docker system prune -f || true'
        }
    }
}
