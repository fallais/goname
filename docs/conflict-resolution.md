# Conflict Resolution in GoName

GoName now includes comprehensive conflict resolution capabilities to handle file naming conflicts during rename operations.

## Conflict Resolution Strategies

### 1. Skip Conflicts (skip)
Skips files that would cause naming conflicts.
```bash
goname apply --conflict skip
```

### 2. Append Number (append) - Default
Adds a number suffix to conflicting files: `movie.mkv` → `movie (1).mkv`
```bash
goname apply --conflict append
```

### 3. Append Timestamp (timestamp)
Adds a timestamp to conflicting files: `movie.mkv` → `movie_20231224_143045.mkv`
```bash
goname apply --conflict timestamp
```

### 4. Prompt User (prompt)
Interactively asks the user what to do for each conflict.
```bash
goname apply --conflict prompt
```

### 5. Overwrite (overwrite)
Overwrites existing files without confirmation.
```bash
goname apply --conflict overwrite
```

### 6. Create Backup (backup)
Creates a backup of existing files before overwriting.
```bash
goname apply --conflict backup --backup-dir ./backups
```

## Usage Examples

### Basic Usage with Conflict Resolution
```bash
# Plan with conflict detection
goname plan --conflict append

# Apply with number appending (default)
goname apply --conflict append

# Apply with interactive prompts
goname apply --conflict prompt
```

### Backup Strategy
```bash
# Create backups in a specific directory
goname apply --conflict backup --backup-dir ./goname-backups

# Create backups in default location (.goname_backups)
goname apply --conflict backup
```

### Batch Processing with Skip Strategy
```bash
# Skip all conflicts and only rename files without conflicts
goname apply --conflict skip
```

## Configuration

Conflict resolution can also be configured via environment variables or config file:

### Environment Variables
```bash
export GONAME_CONFLICT=append
export GONAME_BACKUP_DIR=/path/to/backups
```

### Config File (~/.goname.yaml)
```yaml
conflict: append
backup_dir: /home/user/goname-backups
```

## Conflict Resolution Results

When using conflict resolution, GoName will display:
- ✓ Successfully renamed files
- ⚠ Skipped files due to conflicts
- ✗ Failed operations
- Additional information about the resolution action taken

Example output:
```
Processing [1/3]: movie.2023.mkv
  ✓ Renamed: Movie (2023).mkv (append_number)
  
Processing [2/3]: show.s01e01.mkv  
  ⚠ Skipped (conflict): show.s01e01.mkv

Processing [3/3]: another.movie.mkv
  ✓ Renamed: Another Movie (2023).mkv [backup: Another Movie (2023)_backup_20231224_143045.mkv]
```

## Integration with State Management

Conflict resolution integrates with GoName's state management system:
- All rename operations (including conflict resolutions) are tracked
- You can revert operations that used conflict resolution
- Backup files are tracked and can be restored

```bash
# View all operations including conflict resolutions
goname state ls

# Revert operations (will also handle conflict resolution cleanup)
goname state revert --last
```
