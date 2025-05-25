import subprocess
import re
import os
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

run_directory = os.path.abspath(os.path.join(os.path.dirname(__file__), "../main"))

modes = ["s", "p", "w"]
threads = 1
runs = 3

re_s = re.compile(r"S-TIME: (\d+\.\d+)")
re_p = re.compile(r"P-TIME: (\d+\.\d+)")
re_ws = re.compile(r"WS-TIME: (\d+\.\d+)")

start_pos = "f" # fischer, more move possibilities, better testing
depth = 5

times     = {m: {} for m in modes}
speedups  = {"p": {}, "w": {}}

for mode in modes:
    if mode == "s":
        threads = [1]
    else:
        threads = [2,4,6,8,12]
    
    for num in threads:
        total = 0.0
        for i in range(runs):
            print(".")
            command = ["go", "run", "main.go", mode, str(num), start_pos, str(depth)]
            output = subprocess.check_output(command, cwd=run_directory).decode("utf-8")
            if mode == "s":
                m = re_s.search(output)
                long_mode = "Sequential"
            elif mode == "p":
                m = re_p.search(output)
                long_mode = "Parallel"
            else:
                m = re_ws.search(output)
                long_mode = "Work-Stealing"
            total += float(m.group(1))

        avg = total / runs
        times[mode][num] = avg
        print(f"{long_mode} at {num} threads = {avg:.4f}s")

seq_time = times["s"][1]

for mode in ("p", "w"):
    for thread, time in times[mode].items():
        speedups[mode][thread] = seq_time / time

plt.figure(figsize=(10,6), facecolor='0.95')
ax = plt.gca()
ax.set_facecolor("0.8")

for mode, style, label in [
    ("p", "o-", "Parallel Threads"),
    ("w", "s--", "Work Stealing"),
]:
    xs = sorted(speedups[mode].keys())
    ys = [speedups[mode][x] for x in xs]
    plt.plot(xs, ys, style, label=label)

plt.xlabel("Number of Threads")
plt.ylabel("Speedup")
plt.title("Speedup vs. Sequential")
plt.xticks([1,2,4,6,8,12])
plt.legend()
plt.grid(True)
plt.savefig("speedup.png")
print("Saved plot to speedup.png")
