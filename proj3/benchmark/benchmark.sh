#!/bin/bash
#
#SBATCH --mail-user=smcveigh@cs.uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj3_benchmark 
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/smcveigh/project-3-shanemcv/proj3/benchmark
#SBATCH --partition=fast 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=12:00:00


module load golang/1.19
echo "Running benchmark testing"
python3 -u benchmark.py

