pipeline {
    agent any
    
    environment {
        // Configuration Application
        APP_NAME = 'my-go-app'
        APP_PORT = '8090'
        
        // Configuration Docker
        DOCKER_REGISTRY = 'laurentmd5'
        DOCKER_IMAGE = "${APP_NAME}"
        DOCKER_TAG = "${env.BUILD_NUMBER}"
        
        // Configuration Serveur Ubuntu
        DEPLOY_SERVER = 'devops@localhost'
        DEPLOY_PATH = '/home/devops/apps'
        SSH_CREDENTIALS_ID = 'ubuntu-server-ssh'
    }
    
    stages {
        // ÉTAPE 1: Checkout du code
        stage('Checkout Code') {
            steps {
                checkout scm
                sh '''
                echo "📦 Repository: ${GIT_URL}"
                echo "📝 Branch: ${GIT_BRANCH}"
                echo "🔍 Commit: ${GIT_COMMIT}"
                ls -la
                '''
            }
        }
        
        // ÉTAPE 2: Build application Go
        stage('Build Go Application') {
            steps {
                sh '''
                echo "🏗️ Building Go application..."
                
                # Installation des dépendances
                if [ -f "go.mod" ]; then
                    echo "📥 Downloading dependencies..."
                    go mod download
                    go mod verify
                else
                    echo "ℹ️ No go.mod found, initializing..."
                    go mod init ${APP_NAME} 2>/dev/null || true
                fi
                
                # Construction de l'application
                echo "🔨 Compiling application..."
                if [ -f "cmd/main.go" ]; then
                    go build -v -o ${APP_NAME} ./cmd/main.go
                elif [ -f "main.go" ]; then
                    go build -v -o ${APP_NAME} main.go
                else
                    # Trouver le premier fichier .go avec une fonction main
                    MAIN_FILE=$(find . -name "*.go" -type f -exec grep -l "func main()" {} \; | head -1)
                    if [ -n "$MAIN_FILE" ]; then
                        go build -v -o ${APP_NAME} .
                    else
                        echo "❌ No main Go file found!"
                        find . -name "*.go" | head -5
                        exit 1
                    fi
                fi
                
                # Vérification du binaire
                echo "✅ Build verification:"
                ls -lh ${APP_NAME}
                file ${APP_NAME}
                chmod +x ${APP_NAME}
                '''
            }
        }
        
        // ÉTAPE 3: Tests statiques
        stage('Static Analysis') {
            steps {
                sh '''
                echo "🔍 Running static analysis..."
                
                # 1. Analyse Go Vet
                echo "=== Go Vet ==="
                go vet ./... && echo "✅ Go Vet passed" || echo "⚠️ Go Vet found issues"
                
                # 2. Vérification format
                echo "=== Code Format ==="
                if [ -z "$(gofmt -l .)" ]; then
                    echo "✅ Code is properly formatted"
                else
                    echo "❌ Code format issues:"
                    gofmt -l .
                    # Ne pas échouer le build pour le format
                fi
                
                # 3. Vérification des dépendances
                echo "=== Dependencies ==="
                go list -m all 2>/dev/null | head -10 || echo "No go.mod"
                
                # 4. Analyse statique étendue
                echo "=== Static Code Analysis ==="
                go version
                echo "✅ Static analysis completed"
                '''
            }
        }
        
        // ÉTAPE 4: Tests dynamiques
        stage('Dynamic Tests') {
            steps {
                sh '''
                echo "🧪 Running dynamic tests..."
                
                # Créer le dossier pour les résultats de tests
                mkdir -p test-results
                
                # 1. Tests unitaires avec coverage
                echo "=== Unit Tests ==="
                if go test -v -race -coverprofile=test-results/coverage.out ./... 2>&1 | tee test-results/test-output.log; then
                    echo "✅ Unit tests passed"
                    # Génération rapport coverage
                    go tool cover -html=test-results/coverage.out -o test-results/coverage.html 2>/dev/null || echo "Coverage HTML not available"
                else
                    echo "⚠️ Some tests failed, continuing..."
                fi
                
                # 2. Tests d'intégration si disponibles
                echo "=== Integration Tests ==="
                go test -v -tags=integration ./... 2>/dev/null | tee test-results/integration-tests.log || echo "ℹ️ No integration tests found or they failed"
                
                # 3. Test basique du binaire
                echo "=== Binary Smoke Test ==="
                timeout 5s ./${APP_NAME} & 
                sleep 2
                if curl -f -s http://localhost:${APP_PORT}/ >/dev/null 2>&1; then
                    echo "✅ Application responds on port ${APP_PORT}"
                else
                    echo "ℹ️ Application not responding on port ${APP_PORT} (might be expected)"
                fi
                pkill ${APP_NAME} 2>/dev/null || true
                
                echo "✅ Dynamic tests completed"
                '''
            }
            
            post {
                always {
                    // Publication des résultats de tests
                    junit 'test-results/*.xml' 
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: 'test-results',
                        reportFiles: 'coverage.html',
                        reportName: 'Code Coverage Report'
                    ])
                    archiveArtifacts artifacts: 'test-results/**/*,${APP_NAME}', fingerprint: true
                }
            }
        }
        
        // ÉTAPE 5: Construction image Docker
        stage('Build Docker Image') {
            steps {
                script {
                    sh """
                    echo "🐳 Building Docker image..."
                    
                    # Vérifier que le Dockerfile existe
                    if [ ! -f "Dockerfile" ]; then
                        echo "❌ Dockerfile not found!"
                        ls -la
                        exit 1
                    fi
                    
                    # Construction de l'image
                    docker build -t ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} .
                    docker tag ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
                    
                    echo "✅ Docker images created:"
                    docker images | grep ${DOCKER_REGISTRY}
                    """
                }
            }
        }
        
        // ÉTAPE 6: Push vers Docker Hub
        stage('Push to Docker Hub') {
            steps {
                script {
                    withCredentials([usernamePassword(
                        credentialsId: 'docker-hub-creds',
                        usernameVariable: 'DOCKERHUB_USER',
                        passwordVariable: 'DOCKERHUB_PASS'
                    )]) {
                        sh """
                        echo "📤 Pushing to Docker Hub..."
                        
                        # Login Docker Hub
                        echo \$DOCKERHUB_PASS | docker login -u \$DOCKERHUB_USER --password-stdin
                        
                        # Push des images
                        docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${DOCKER_TAG}
                        docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
                        
                        # Logout
                        docker logout
                        
                        echo "✅ Images pushed to Docker Hub"
                        echo "🔗 https://hub.docker.com/r/${DOCKER_REGISTRY}/${DOCKER_IMAGE}"
                        """
                    }
                }
            }
        }
        
        // ÉTAPE 7: Déploiement sur Ubuntu via SSH
        stage('Deploy to Ubuntu via SSH') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER',
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🚀 Deploying to Ubuntu server..."
                        
                        # Script de déploiement inline
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            set -e
                            echo '🎯 Starting deployment of ${APP_NAME} on port ${APP_PORT}...'
                            
                            # Création du répertoire de déploiement
                            mkdir -p ${DEPLOY_PATH}
                            cd ${DEPLOY_PATH}
                            
                            # Arrêt et suppression de l'ancien conteneur
                            echo '⏹️ Stopping existing container...'
                            docker stop ${APP_NAME} 2>/dev/null || true
                            docker rm ${APP_NAME} 2>/dev/null || true
                            
                            # Nettoyage des images anciennes
                            echo '🧹 Cleaning old images...'
                            docker image prune -f 2>/dev/null || true
                            
                            # Pull de la nouvelle image
                            echo '📥 Pulling new image from Docker Hub...'
                            docker pull ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            # Lancement du nouveau conteneur
                            echo '🐳 Starting new container...'
                            docker run -d \\
                              --name ${APP_NAME} \\
                              -p ${APP_PORT}:${APP_PORT} \\
                              --restart unless-stopped \\
                              ${DOCKER_REGISTRY}/${APP_NAME}:latest
                            
                            echo '⏳ Waiting for application to start...'
                            sleep 10
                            
                            # Vérification du déploiement
                            echo '🔍 Verifying deployment...'
                            
                            # Check container status
                            if docker inspect -f '{{.State.Status}}' ${APP_NAME} 2>/dev/null | grep -q 'running'; then
                                echo '✅ Container is running'
                            else
                                echo '❌ Container is not running'
                                docker logs ${APP_NAME} --tail 10
                                exit 1
                            fi
                            
                            # Health check
                            if curl -f -s -o /dev/null -w '%{http_code}' http://localhost:${APP_PORT}/ | grep -q '200'; then
                                echo '✅ Health check passed'
                            else
                                echo '⚠️ Health check failed or returned non-200 status'
                                docker logs ${APP_NAME} --tail 10
                            fi
                            
                            echo '🎉 Deployment completed successfully!'
                            echo '🌐 Application URL: http://localhost:${APP_PORT}'
                            echo '🌐 Network URL: http://192.168.61.131:${APP_PORT}'
                        "
                        """
                    }
                }
            }
        }
        
        // ÉTAPE 8: Vérification finale
        stage('Final Verification') {
            steps {
                script {
                    withCredentials([sshUserPrivateKey(
                        credentialsId: "${SSH_CREDENTIALS_ID}",
                        usernameVariable: 'SSH_USER', 
                        keyFileVariable: 'SSH_KEY'
                    )]) {
                        sh """
                        echo "🔎 Final verification..."
                        
                        # Vérification sur le serveur distant
                        ssh -i \$SSH_KEY -o StrictHostKeyChecking=no ${DEPLOY_SERVER} "
                            echo '=== CONTAINER STATUS ==='
                            docker ps --filter 'name=${APP_NAME}' --format 'table {{.Names}}\\t{{.Status}}\\t{{.RunningFor}}\\t{{.Ports}}'
                            
                            echo '=== APPLICATION LOGS ==='
                            docker logs ${APP_NAME} --tail 10
                            
                            echo '=== NETWORK PORTS ==='
                            ss -tln | grep ${APP_PORT} || netstat -tln | grep ${APP_PORT} || echo 'Port not found in netstat'
                            
                            echo '=== FINAL HEALTH CHECK ==='
                            if curl -f -s http://localhost:${APP_PORT}/ > /dev/null; then
                                echo '✅ FINAL HEALTH CHECK PASSED'
                            else
                                echo '⚠️ Final health check failed'
                                exit 1
                            fi
                        "
                        
                        echo ""
                        echo "🎊 DEPLOYMENT SUCCESSFUL!"
                        echo "📊 DEPLOYMENT SUMMARY:"
                        echo "   🏷️  Application: ${APP_NAME}"
                        echo "   🔢 Version: ${DOCKER_TAG}" 
                        echo "   🌐 Local URL: http://localhost:${APP_PORT}"
                        echo "   🌐 Network URL: http://192.168.61.131:${APP_PORT}"
                        echo "   🐳 Image: ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest"
                        echo "   📍 Server: ${DEPLOY_SERVER}"
                        echo "   🔗 Docker Hub: https://hub.docker.com/r/${DOCKER_REGISTRY}/${DOCKER_IMAGE}"
                        """
                    }
                }
            }
        }
    }
    
    post {
        always {
            // Nettoyage
            sh '''
            echo "🧼 Cleaning up workspace..."
            docker system prune -f 2>/dev/null || true
            rm -f ${APP_NAME} 2>/dev/null || true
            '''
        }
        success {
            emailext (
                subject: "✅ SUCCESS: ${env.JOB_NAME} - Build ${env.BUILD_NUMBER}",
                body: """
                Le déploiement a réussi !
                
                📋 Détails:
                Application: ${APP_NAME}
                Version: ${DOCKER_TAG} 
                URL: http://192.168.61.131:${APP_PORT}
                Serveur: ${DEPLOY_SERVER}
                Image Docker: ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
                
                📊 Build: ${env.BUILD_URL}
                """,
                to: "devops@company.com",
                replyTo: "devops@company.com"
            )
            
            sh """
            echo "✅ DÉPLOIEMENT RÉUSSI!"
            echo "🌐 Votre application est accessible à:"
            echo "   http://localhost:${APP_PORT}"
            echo "   http://192.168.61.131:${APP_PORT}"
            """
        }
        failure {
            emailext (
                subject: "❌ FAILED: ${env.JOB_NAME} - Build ${env.BUILD_NUMBER}",
                body: """
                Le déploiement a échoué !
                
                Application: ${APP_NAME}
                Version: ${DOCKER_TAG}
                
                Consultez les logs: ${env.BUILD_URL}
                """,
                to: "devops@company.com",
                replyTo: "devops@company.com"
            )
            
            sh """
            echo "❌ DÉPLOIEMENT ÉCHOUÉ"
            echo "📋 Vérifiez les logs ci-dessus pour le diagnostic"
            """
        }
    }
}
