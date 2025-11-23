# Building Custom wrk2 with Continuous Mode

This guide explains how to modify wrk2 to support native continuous/infinite mode and use it with the load generator.

## Why Native Continuous Mode?

The default implementation uses a shell loop to run wrk2 continuously. For some use cases, you may want wrk2 itself to support infinite running without the shell wrapper. This provides:

- Single wrk2 process instead of repeated restarts
- Continuous statistics without cycle breaks
- Native signal handling and control

## Modifying wrk2 Source Code

### Option 1: Add Infinite Duration Support

Modify wrk2 to accept `0` or `-1` as duration to run infinitely.

1. Clone wrk2:
```bash
cd /tmp
git clone https://github.com/giltene/wrk2.git
cd wrk2
```

2. Edit `src/wrk.c` - find the `main()` function and locate the duration parsing section (around line 450-460):

```c
// Find this section:
if (duration.tv_sec || duration.tv_usec) {
    timeout = &duration;
}

// Modify to add infinite mode check:
if (duration.tv_sec == 0 && duration.tv_usec == 0) {
    // Duration was explicitly set to 0, run infinitely
    timeout = NULL;
    cfg.duration = UINT64_MAX; // Set to max value for infinite
} else if (duration.tv_sec || duration.tv_usec) {
    timeout = &duration;
}
```

3. Modify the stop condition in the main loop (around line 580-600):

```c
// Find the time check in main loop:
if (cfg.duration && now >= stop_at) break;

// Change to handle infinite duration:
if (cfg.duration != UINT64_MAX && now >= stop_at) break;
```

### Option 2: Add a New `-i` (infinite) Flag

Add a command-line flag for infinite mode:

1. In `src/wrk.c`, add to the usage string (around line 70):

```c
static const char *USAGE =
    "Usage: wrk <options> <url>                            \n"
    "  Options:                                            \n"
    ...
    "    -d, --duration  <T>  Duration of test              \n"
    "    -i, --infinite       Run test infinitely           \n"  // ADD THIS
    ...
```

2. Add infinite flag variable (around line 110):

```c
static struct config {
    ...
    bool infinite;  // ADD THIS
} cfg;
```

3. Add flag parsing (in the getopt section, around line 450):

```c
case 'i':
    cfg.infinite = true;
    break;
```

4. Modify the main loop condition:

```c
// In main loop:
if (!cfg.infinite && cfg.duration && now >= stop_at) break;
```

## Building the Custom Binary

After making your modifications:

```bash
# Build wrk2
cd wrk2
make clean
make -j$(nproc)

# The binary is at: wrk
# Test it:
./wrk --help
```

## Using with DeathStarBench Load Generator

### 1. Create wrk2-custom directory

In each service directory (hotelReservation, socialNetwork, mediaMicroservices):

```bash
cd hotelReservation  # or socialNetwork or mediaMicroservices
mkdir -p wrk2-custom
```

### 2. Copy your custom wrk2 binary

```bash
cp /tmp/wrk2/wrk wrk2-custom/wrk
chmod +x wrk2-custom/wrk
```

### 3. Build Docker image with custom binary

```bash
# For hotel reservation:
cd hotelReservation
docker build -f Dockerfile-loader-custom -t 10.10.10.240/library/hotelreservation-loader:custom .
docker push 10.10.10.240/library/hotelreservation-loader:custom

# For social network:
cd socialNetwork
docker build -f Dockerfile-loader-custom -t 10.10.10.240/library/socialnetwork-loader:custom .
docker push 10.10.10.240/library/socialnetwork-loader:custom

# For media microservices:
cd mediaMicroservices
docker build -f Dockerfile-loader-custom -t 10.10.10.240/library/mediamicroservices-loader:custom .
docker push 10.10.10.240/library/mediamicroservices-loader:custom
```

### 4. Update Helm values to use custom image

```yaml
# In values.yaml:
container:
  image: 10.10.10.240/library/hotelreservation-loader
  imageVersion: "custom"

loadTest:
  # Use native infinite mode if you added -i flag:
  additionalArgs: "-D exp -L -i"
  # Or use duration 0 if you used Option 1:
  duration: "0"
```

### 5. Update deployment template (optional)

If using native continuous mode, you can simplify the deployment template to remove the shell loop:

```yaml
# In templates/deployment.yaml, replace the command section with:
command:
- /bin/sh
- -c
- |
  echo "Waiting for service to be ready..."
  until nc -z {{ .Values.loadTest.targetHost }} {{ .Values.loadTest.targetPort }}; do
    echo "Waiting..."
    sleep 5
  done
  echo "Service ready. Starting continuous load test..."
  wrk {{ .Values.loadTest.additionalArgs }} \
    -t {{ .Values.loadTest.threads }} \
    -c {{ .Values.loadTest.connections }} \
    -s /hotel-reservation/wrk2/scripts/hotel-reservation/{{ .Values.loadTest.script }} \
    {{ .Values.loadTest.targetUrl }} \
    -R {{ .Values.loadTest.rate }}
```

## Example Modifications

### Simple Infinite Duration Patch

Create a file `wrk2-infinite.patch`:

```patch
--- a/src/wrk.c
+++ b/src/wrk.c
@@ -455,7 +455,11 @@ int main(int argc, char **argv) {
     
     duration = cfg.duration ? cfg.duration : 10;
     
-    timeout = &duration;
+    if (duration == 0) {
+        // Infinite mode
+        timeout = NULL;
+    } else {
+        timeout = &duration;
+    }
     
     stop_at = time_us() + (cfg.duration * 1000000);
```

Apply with:
```bash
cd wrk2
patch -p1 < wrk2-infinite.patch
make clean && make
```

## Testing Your Custom Binary

```bash
# Test normal mode (10 second test):
./wrk -t2 -c10 -d10s -R100 http://localhost:8080

# Test infinite mode with -d0 (if you implemented Option 1):
./wrk -t2 -c10 -d0 -R100 http://localhost:8080

# Test infinite mode with -i flag (if you implemented Option 2):
./wrk -t2 -c10 -i -R100 http://localhost:8080
```

Press Ctrl+C to stop the infinite test.

## Notes

- The shell loop approach (`continuous: true` in values.yaml) still works and doesn't require custom wrk2
- Native continuous mode may have different statistics reporting behavior
- Consider adding signal handling for graceful shutdown in your wrk2 modifications
- The custom binary approach requires you to maintain the binary and rebuild when needed

## Troubleshooting

**Binary not found in Docker image:**
- Ensure `wrk2-custom/wrk` exists before building
- Check file permissions: `chmod +x wrk2-custom/wrk`

**Segmentation fault:**
- Ensure wrk2 was built with compatible libraries
- Check that all dependencies are installed in the Docker image

**Load test doesn't run infinitely:**
- Verify your code modifications are correct
- Test the binary locally before dockerizing
- Check logs: `kubectl logs -n namespace pod-name`
